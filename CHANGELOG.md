# Changelog

Все значимые изменения в проекте описываются здесь. Формат — [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/), версионирование — [SemVer](https://semver.org/lang/ru/).

## [0.2.0] — 2026-05-01

### Главное

`wtf` превратился из «объяснялки последнего вывода» в **терминальный sysadmin-агент**: ты пишешь `wtf <вопрос словами>` — агент сам выполняет диагностические команды на твоей машине, читает их выводы, итерирует и предлагает решение.

### Добавлено

- **Tool-use агент** в `internal/agent/` — итеративный цикл AI ↔ инструменты, до 15 раундов на сессию
- **Безопасный classifier команд** в `internal/exec/classify.go` — 80+ read-only утилит в whitelist, blacklist destructive (sudo, rm, restart, install, push, force, опасные паттерны)
- **Tool-use API всех трёх провайдеров** (Claude / OpenAI / Gemini) вместо парсинга текста — надёжная структура вызовов
- **Память между сессиями** (`internal/memory/`): `~/.wtf/memory/store.json` с TTL по типам записей (machine_fact / service_state / user_preference / resolved_issue), AI-консолидация раз в 20 сессий
- **Auto-retry при rate-limit (429)**: парс `Retry-After`, `x-ratelimit-reset-tokens`, текста "try again in 6.7s"; до 3 попыток с задержкой
- **Throttle между AI-раундами** (800мс) + **trim истории** — последние 4 раунда полностью, старые tool_result сжимаются до маркера
- **Quiet UI с timestamps**: каждая строка `[HH:MM:SS]`, спиннер в позиции финальной иконки (без прыжков), один-line summary команды
- **Verbose режим**: `wtf -v` показывает полный вывод команд; при `exit≠0` в обычном режиме — последние 5 строк автоматически
- **Graceful Ctrl+C**: `signal.NotifyContext`, память сохраняется при прерывании, exit code 130
- **Versioning памяти**: поле `version` в `store.json`, апгрейд старых файлов автоматический, защита от форматов из будущего
- **Детект дублей команд**: модель не может запустить одну команду дважды в сессии (защита от зацикливания)
- **Pipe-режим**: `cat err.log | wtf что не так` — содержимое stdin как контекст
- **Команды управления памятью**: `wtf memory show`, `wtf memory clear`

### Удалено

- `internal/shellhook/` — DEBUG-trap, last_meta, last_output больше не нужны
- `internal/cache/` — для интерактивного агента кэш бесполезен
- Команды: `wtf init`, `wtfc <cmd>`, `wtf --rerun`, `wtf --explain "..."`, `--no-cache`
- Все хуки в .bashrc/.zshrc/fish/PowerShell

### Изменено

- `cmd/wtf/main.go` полностью переписан под новый UX: позиционный парсинг (`wtf nginx не стартует` без кавычек), интерактивный fallback при пустом вопросе
- `provider.Request → provider.ChatRequest`: теперь передаются messages + tools + system, поддерживается tool-use
- README и frontend-лендинг полностью обновлены под новую концепцию

## [0.1.0] — 2026-04-30

Первый публичный релиз.

### Добавлено

- CLI на Go: один кросс-платформенный бинарник (~8 МБ) для macOS / Linux / Windows
- Три AI-провайдера на выбор: **Claude** (Anthropic), **OpenAI**, **Gemini** (Google) — переключение через флаг или конфиг
- Shell-хуки для **bash**, **zsh**, **fish**, **PowerShell** (`wtf init`)
- Fallback `wtf --rerun` — перезапуск последней команды с захватом stdout/stderr (если хук не установлен)
- Явный режим `wtf --explain "<text>"`
- Локальный кеш в `~/.wtf/cache/` с TTL 30 дней (sha256 по provider+model+lang+command+output)
- Чистка секретов перед отправкой — 13 regex-правил (sk-/sk-ant-/AIza/gh*_/xox*-, JWT, Bearer, AWS, private keys, password=, basic-auth URL, email, $HOME → ~)
- Consent-баннер при первом запуске с полным списком отправляемых полей
- Spinner с Braille-frames в стиле Claude CLI
- Цветной Markdown-рендер ответа (заголовки, инлайн-код, code-блоки, bold)
- Интерактивный wizard `wtf config` для первой настройки
- Автоматический запуск wizard при первом запуске если ни один ключ не задан
- Сбор контекста: OS, shell, cwd, git branch, package manager (по lock-файлам)
- Поддержка ENV-переменных как fallback для ключей (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`/`GOOGLE_API_KEY`)
- Prompt caching для Claude (system-промпт кешируется)
- RU + EN языки ответа

### Известные ограничения

- Нет streaming ответа — ждём весь ответ одним куском
- Нет `--offline` режима через ollama (в roadmap)
- Нет `wtf --share` через GitHub Gist (в roadmap)
- На Windows только PowerShell (cmd.exe не поддерживается)
