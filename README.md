# wtf — терминальный sysadmin-агент

> Опиши проблему словами — `wtf nginx не стартует` — агент сам выполнит диагностические команды на твоей машине и вернёт решение. Запоминает что узнал о тебе и твоём сервере между запусками. Один Go-бинарь, три AI-провайдера на выбор, MIT.

[![go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platforms](https://img.shields.io/badge/platforms-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)]()
[![Providers](https://img.shields.io/badge/AI-Claude%20%7C%20OpenAI%20%7C%20Gemini-fbbf24)]()

## Установка

### macOS / Linux

```bash
curl -sSL https://raw.githubusercontent.com/kitay-sudo/wtf/main/install.sh | sudo bash
```

или через Homebrew:

```bash
brew install kitay-sudo/wtf/wtf
```

### Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/kitay-sudo/wtf/main/install.ps1 | iex
```

### Из исходников

```bash
git clone https://github.com/kitay-sudo/wtf
cd wtf
go build -o wtf ./cmd/wtf
sudo mv wtf /usr/local/bin/
```

## Что это

`wtf` — это терминальный агент. Ты пишешь ему словами что у тебя сломалось — он сам выполняет на твоей машине безопасные read-only команды (`systemctl status`, `journalctl`, `ls`, `cat`, и т.д.), читает их вывод, и итеративно докапывается до причины. Когда нужно что-то поменять (sudo, restart, install, rm) — показывает точную команду, чтобы ты выполнил сам. Никогда не запускает destructive-команды без твоего ведома.

Между сессиями агент запоминает что узнал о твоей машине: версии сервисов, нестандартные пути конфигов, твои предпочтения. В следующий раз он уже знает контекст.

### Пример

```
$ wtf nginx не стартует

  → проверяю статус сервиса
  $ systemctl status nginx -l
  │ ● nginx.service - A high performance web server
  │      Active: failed (Result: exit-code) since Thu 2026-04-30 18:15
  │     Process: 2418961 ExecStart=/usr/sbin/nginx ... (code=exited, status=1)
  │ nginx: [emerg] unknown directive "serer_name" in /etc/nginx/sites-enabled/default:5
  └ exit=3 · 142ms

  → читаю последние логи
  $ journalctl -u nginx -n 20 --no-pager
  │ apr 30 18:15:08 nginx[2418961]: nginx: [emerg] unknown directive "serer_name"
  │ apr 30 18:15:08 nginx[2418961]: nginx: configuration file /etc/nginx/nginx.conf test failed
  └ exit=0 · 89ms

  Проблема: опечатка в /etc/nginx/sites-enabled/default:5 — `serer_name`
  вместо `server_name`.

  ⚠ выполни сам (требует sudo / меняет систему):
    что эта команда сделает: исправит опечатку и перезапустит nginx
    $ sudo sed -i 's/serer_name/server_name/' /etc/nginx/sites-enabled/default
    $ sudo systemctl restart nginx
```

### Ещё пример — пайп логов

```bash
$ tail -f /var/log/app.log | head -100 | wtf что не так
```

Агент возьмёт переданный лог как контекст, при необходимости запустит дополнительные команды для диагностики, вернёт решение.

## Возможности

- 🤖 **Tool-use агент** — модель сама решает какие команды запустить, читает их вывод, итерирует. До 15 раундов диагностики на сессию.
- 🛡️ **Safe-by-default** — встроенный classifier разделяет команды на безопасные (run_command — выполняем сами) и destructive (show_command — показываем юзеру). sudo/rm/restart/install никогда не запускаются автоматически.
- 🧠 **Память между сессиями** — после каждой решённой проблемы агент сохраняет ключевые факты: версии сервисов, пути конфигов, предпочтения. В `~/.wtf/memory/store.json`. Раз в N сессий сжимается через AI чтобы не разрослось.
- 🌐 **Три провайдера** — Claude, OpenAI, Gemini. Все три используют tool-use API для надёжного парсинга. Переключение через `--provider`.
- 🔒 **Чистка секретов** — 13 regex-правил перед отправкой в API и перед записью в память. Токены, JWT, пароли, email, basic-auth URL — вычищаются.
- 🪶 **Один бинарь** — Go, ~8 МБ, без зависимостей. Никаких shell-хуков, обёрток, изменений в .bashrc.
- 🌍 **RU + EN** — `wtf --lang en` если нужен английский.

## Быстрый старт

```bash
# 1. Настрой провайдера (интерактивный wizard)
wtf config

# 2. Запусти диагностику — словами, без кавычек
wtf nginx не стартует
wtf почему медленно работает диск
wtf что значит этот вывод | cat strange-error.log

# 3. Посмотри что агент о тебе помнит (опционально)
wtf memory show
```

## Команды

```
wtf <вопрос>                  запустить диагностику
cat err.log | wtf <вопрос>    диагностика + содержимое stdin как контекст
wtf                           интерактив: запросит вопрос

  --provider <name>           разово выбрать провайдера (claude|openai|gemini)
  --lang <ru|en>              язык ответа

wtf config                    настройка провайдера, ключей, моделей
wtf config show               показать текущий конфиг
wtf config set k=v            задать значение

wtf memory show               показать что агент о тебе помнит
wtf memory clear              стереть память

wtf version                   версия
```

## Конфигурация

Конфиг лежит в `~/.wtf/config.json` (mode 0600):

```json
{
  "default_provider": "claude",
  "language": "ru",
  "providers": {
    "claude": { "api_key": "sk-ant-...", "model": "claude-haiku-4-5-20251001" },
    "openai": { "api_key": "sk-...",     "model": "gpt-4o-mini" },
    "gemini": { "api_key": "AIza...",    "model": "gemini-2.0-flash" }
  }
}
```

Можно задавать ключи через переменные окружения — конфиг переопределяет:

- `ANTHROPIC_API_KEY` — для Claude
- `OPENAI_API_KEY` — для OpenAI
- `GEMINI_API_KEY` или `GOOGLE_API_KEY` — для Gemini

## Как это работает

```
   wtf <вопрос>
       │
       ├──► собираем контекст: OS, shell, cwd, git, package manager
       ├──► загружаем память из ~/.wtf/memory/store.json
       │
       ▼
   ┌─────────────────── цикл агента (до 15 итераций) ───────────────────┐
   │                                                                    │
   │  AI → решает что нужно, вызывает tool:                             │
   │                                                                    │
   │   • run_command(cmd, reason) — wtf классифицирует команду:         │
   │       safe        → запускает на машине, возвращает stdout/stderr  │
   │       destructive → отказывает, говорит "используй show_command"   │
   │       unknown     → отказывает                                     │
   │                                                                    │
   │   • show_command(cmd, reason) — wtf печатает юзеру с пометкой ⚠   │
   │       юзер выполняет сам, вывод (если нужно) пейстит обратно       │
   │                                                                    │
   │   • finish(summary, notes) — финал.                                │
   │       summary  → в терминал юзеру (Markdown)                       │
   │       notes    → в ~/.wtf/memory/store.json после редакции         │
   │                                                                    │
   └────────────────────────────────────────────────────────────────────┘
       │
       ├──► раз в 20 сессий — консолидация памяти через AI (сжимает старое)
       └──► выход
```

### Безопасность и классификация команд

[`internal/exec/classify.go`](internal/exec/classify.go) разделяет команды на три класса:

- **safe** — read-only утилиты: `ls cat tail grep ps systemctl status journalctl docker ps git status nginx -t apt list ip ss lsof df` и т.д. Whitelist первого токена + проверка subcommand (например `git push` — destructive даже если `git` в whitelist).
- **destructive** — `sudo rm mv dd mkfs chmod chown systemctl restart apt install pip install git push docker rm ...` плюс опасные паттерны (`> /etc/...`, `| sh`, `&` в фоне).
- **unknown** — всё остальное. Отказываем по умолчанию.

Безопасные команды запускаются автоматически с лимитом 30s и обрезкой вывода до 8 КБ. Destructive — только показываются юзеру.

### Память

После сессии агент сохраняет короткие заметки 4 типов:

- `machine_fact` — стабильные факты (TTL 0 = вечно)
- `service_state` — версии и состояние сервисов (TTL 30 дней)
- `user_preference` — привычки юзера (TTL 0)
- `resolved_issue` — решённые проблемы (TTL 30 дней)

Все записи прогоняются через `internal/redact/` — секреты не уезжают в файл. При старте загружается до 2 КБ контекста в system-промпт.

Раз в 20 сессий (или при превышении 100 записей) запускается консолидация: AI получает все заметки, возвращает сжатый список из самых ценных, старые перезаписываются.

### Redaction

Перед отправкой в API и перед записью в память работают regex-фильтры:

| Класс | Шаблон |
|---|---|
| Anthropic API key | `sk-ant-[A-Za-z0-9_\-]{20,}` |
| OpenAI / generic API key | `sk-[A-Za-z0-9_\-]{20,}` |
| Google API key | `AIza[0-9A-Za-z_\-]{35}` |
| GitHub token | `gh[pousr]_[A-Za-z0-9]{20,}` |
| Slack token | `xox[baprs]-...` |
| JWT | `eyJ...\.eyJ...\....` |
| Bearer token | `Bearer <anything>` |
| AWS access key / secret | `AKIA...` / `aws_secret*=...` |
| Private keys | `-----BEGIN ... PRIVATE KEY-----...` |
| Generic password/secret/token KV | `password=...`, `token=...` |
| Basic-auth URL | `https://user:pass@host` |
| Email | `user@host.tld` |
| Home path | `$HOME → ~` |

## Roadmap

- [ ] `--offline` режим через ollama (локальный LLM)
- [ ] Streaming tool-use ответов
- [ ] Многосессионность: возврат к прошлой проблеме через `wtf resume`
- [ ] Brew tap, scoop, AUR, deb/rpm

## Безопасность

- Конфиг и память лежат в `~/.wtf/` с правами `0600`.
- Ключи отображаются маскированно (`sk-a…XXXX`).
- Никакой удалённой телеметрии.
- Если найдёшь утечку секрета через regex-фильтр — присылай PR с тестом.

## Вклад

Если проект помог — поставь ⭐ на [github.com/kitay-sudo/wtf](https://github.com/kitay-sudo/wtf).

PR welcome. Запустить локально:

```bash
git clone https://github.com/kitay-sudo/wtf
cd wtf
go build -o wtf ./cmd/wtf
./wtf "что значит exit 137"
```

Фронтенд лендинга:

```bash
cd frontend
npm install
npm run dev
```

### Поддержать криптой

| Сеть | Адрес |
|---|---|
| **USDT** · TRON (TRC20) | `TF9F2FPkreHVfbe8tZtn4V76j3jLo4SeXM` |
| **TON** · The Open Network | `UQBl88kXWJWyHkDPkWNYQwwSCiCAIfA2DiExtZElwJFlIc1o` |

После доната напиши в Telegram [@kitay9](https://t.me/kitay9) свой ник, добавлю в стену чести на лендинге.

### Рекомендую: Timeweb Cloud

[Timeweb Cloud](https://timeweb.cloud/?i=104289) — российский VPS-хостинг, на котором живут наши боевые сервера. Быстрая панель, NVMe, развёртывание за минуту. *Ссылка партнёрская.*

### Выпуск нового релиза

Релиз — это git-тег `vX.Y.Z`:

```bash
# Windows
scripts\release.bat patch    # bug-fix:    v0.1.0 → v0.1.1
scripts\release.bat minor    # фича:        v0.1.0 → v0.2.0
scripts\release.bat major    # breaking:    v0.1.0 → v1.0.0

# macOS / Linux / Git Bash
./scripts/release.sh patch
```

После пуша тега запускается [`.github/workflows/release.yml`](.github/workflows/release.yml) — matrix-сборка под 6 платформ, архивы, SHA256SUMS, GitHub Release.

## Лицензия

MIT © [kitay-sudo](https://github.com/kitay-sudo)
