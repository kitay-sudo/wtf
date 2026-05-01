// Package memory — долговременная память агента.
//
// Структура файлов:
//
//	~/.wtf/memory/
//	  store.json        — все записи: machine_fact / service_state / user_preference / resolved_issue
//	  consolidated_at   — timestamp последней консолидации (для решения когда сжимать)
//
// Записи имеют TTL и автоматически отбрасываются при загрузке если устарели.
// Перед записью прогоняем через redact, чтобы пароли/токены не уехали в файл.
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bitcoff/wtf/internal/config"
	"github.com/bitcoff/wtf/internal/redact"
)

type Type string

const (
	TypeMachineFact     Type = "machine_fact"
	TypeServiceState    Type = "service_state"
	TypeUserPreference  Type = "user_preference"
	TypeResolvedIssue   Type = "resolved_issue"
)

type Entry struct {
	Type      Type      `json:"type"`
	Key       string    `json:"key"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	TTLDays   int       `json:"ttl_days"` // 0 = вечно
}

// CurrentSchemaVersion — версия структуры store.json. Инкрементируется при
// несовместимых изменениях схемы (новые обязательные поля, удалённые типы).
// Loader умеет читать любую версию ≤ CurrentSchemaVersion и при необходимости
// мигрировать; неизвестную (выше) — отказывается читать с понятной ошибкой.
const CurrentSchemaVersion = 1

// Store — все записи памяти. Грузится в начале сессии, сохраняется в конце.
type Store struct {
	// Version — версия схемы. 0 у старых файлов = "до introducing версионирования",
	// автоматически апгрейдится до 1 при первом сохранении.
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`

	// Хвост запусков агента — для статистики и решения когда консолидировать.
	// Не передаётся в AI, только метаданные.
	SessionCount   int       `json:"session_count"`
	ConsolidatedAt time.Time `json:"consolidated_at"`

	dir string
}

// MaxEntries — после этого порога зовём консолидатор сжать старое.
const MaxEntries = 100

// SessionsBetweenConsolidation — каждые N сессий запускаем консолидатор.
const SessionsBetweenConsolidation = 20

func dir() (string, error) {
	cfgDir, err := config.Dir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(cfgDir, "memory")
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	return d, nil
}

func storePath() (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "store.json"), nil
}

// Load читает store.json и отбрасывает протухшие по TTL записи.
// Если файла нет — возвращает пустой Store без ошибки.
func Load() (*Store, error) {
	d, err := dir()
	if err != nil {
		return nil, err
	}
	p := filepath.Join(d, "store.json")
	store := &Store{dir: d}

	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, store); err != nil {
		return nil, fmt.Errorf("memory: parse store: %w", err)
	}
	store.dir = d

	// Будущая версия (юзер откатился на старый wtf) — не пытаемся понять,
	// чтобы случайно не повредить новый формат. Отдаём пустой Store и пишем
	// предупреждение через ошибку — main.go его покажет.
	if store.Version > CurrentSchemaVersion {
		empty := &Store{Version: CurrentSchemaVersion, dir: d}
		return empty, fmt.Errorf("memory: store.json версии %d не поддерживается (текущая %d) — память пропускается",
			store.Version, CurrentSchemaVersion)
	}
	// Старые файлы (Version == 0) принимаем как есть; апгрейд произойдёт при Save.

	store.dropExpired()
	return store, nil
}

// Save пишет store на диск. Атомарно через temp+rename, mode 0600 (могут быть
// IP/hostname клиента — не публичная инфа).
func (s *Store) Save() error {
	if s.dir == "" {
		d, err := dir()
		if err != nil {
			return err
		}
		s.dir = d
	}
	// Апгрейд старых файлов: при сохранении проставляем актуальную версию.
	if s.Version < CurrentSchemaVersion {
		s.Version = CurrentSchemaVersion
	}
	p := filepath.Join(s.dir, "store.json")
	tmp := p + ".tmp"

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// Add вставляет/обновляет запись по (Type, Key) — если уже есть с таким ключом,
// перезаписываем (это и есть дедуп). Перед записью прогоняем content через redact.
func (s *Store) Add(e Entry) {
	e.Content = redact.Apply(e.Content).Text
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	for i, existing := range s.Entries {
		if existing.Type == e.Type && existing.Key == e.Key {
			s.Entries[i] = e
			return
		}
	}
	s.Entries = append(s.Entries, e)
}

// dropExpired удаляет записи у которых истёк TTL.
func (s *Store) dropExpired() {
	now := time.Now()
	kept := s.Entries[:0]
	for _, e := range s.Entries {
		if e.TTLDays == 0 {
			kept = append(kept, e)
			continue
		}
		if now.Sub(e.CreatedAt) < time.Duration(e.TTLDays)*24*time.Hour {
			kept = append(kept, e)
		}
	}
	s.Entries = kept
}

// SystemContext возвращает текстовое представление памяти для system-промпта.
// Группирует по типам, упорядочивает свежие сверху, обрезает по budget символов.
func (s *Store) SystemContext(budget int) string {
	if len(s.Entries) == 0 {
		return ""
	}
	groups := map[Type][]Entry{}
	for _, e := range s.Entries {
		groups[e.Type] = append(groups[e.Type], e)
	}
	for t := range groups {
		sort.Slice(groups[t], func(i, j int) bool {
			return groups[t][i].CreatedAt.After(groups[t][j].CreatedAt)
		})
	}

	var b strings.Builder
	b.WriteString("Память о пользователе и его машине (используй только релевантное):\n\n")

	order := []struct {
		t     Type
		title string
	}{
		{TypeMachineFact, "Машина"},
		{TypeServiceState, "Сервисы"},
		{TypeUserPreference, "Предпочтения"},
		{TypeResolvedIssue, "История проблем (последние)"},
	}
	for _, g := range order {
		entries := groups[g.t]
		if len(entries) == 0 {
			continue
		}
		fmt.Fprintf(&b, "%s:\n", g.title)
		for _, e := range entries {
			fmt.Fprintf(&b, "  - %s: %s\n", e.Key, e.Content)
		}
		b.WriteString("\n")
	}

	out := b.String()
	if budget > 0 && len(out) > budget {
		// Обрезаем хвост — там самые старые resolved_issue, наименее ценные.
		out = out[:budget] + "\n[...обрезано...]\n"
	}
	return out
}

// MarkSession инкрементит счётчик и проверяет нужна ли консолидация.
// Возвращает true если пора консолидировать (агент вызовет AI для сжатия).
func (s *Store) MarkSession() bool {
	s.SessionCount++
	if len(s.Entries) >= MaxEntries {
		return true
	}
	if s.SessionCount > 0 && s.SessionCount%SessionsBetweenConsolidation == 0 {
		return true
	}
	return false
}

// Replace полностью перезаписывает Entries (используется консолидатором,
// который возвращает сжатый список взамен старого).
func (s *Store) Replace(entries []Entry) {
	for i := range entries {
		entries[i].Content = redact.Apply(entries[i].Content).Text
		if entries[i].CreatedAt.IsZero() {
			entries[i].CreatedAt = time.Now()
		}
	}
	s.Entries = entries
	s.ConsolidatedAt = time.Now()
}

// AsJSON возвращает все записи в JSON-формате — для передачи консолидатору.
func (s *Store) AsJSON() string {
	data, _ := json.MarshalIndent(s.Entries, "", "  ")
	return string(data)
}
