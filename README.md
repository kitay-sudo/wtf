# wtf 🤬 — AI-объяснялка ошибок в терминале

> Падение в терминале — пиши `wtf`. Получаешь короткое объяснение и 2-3 готовых фикса. Один Go-бинарь, три AI-провайдера на выбор, MIT.

[![go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platforms](https://img.shields.io/badge/platforms-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)]()
[![Providers](https://img.shields.io/badge/AI-Claude%20%7C%20OpenAI%20%7C%20Gemini-fbbf24)]()

## Что это

`wtf` — это CLI-утилита, которая объясняет последнюю ошибку в твоём терминале человеческими словами и предлагает 2-3 варианта починить. Не нужно копировать stack trace в Google, переключаться на ChatGPT, выдирать из вывода нужное. Упал — написал `wtf` — получил ответ.

```
$ npm run build
Error: Cannot find module './pages/Landing' imported from src/App.jsx

$ wtf
  ⠹  анализ через claude (haiku-4-5)...
  ✓  ответ от claude

  Что случилось: Vite не находит файл src/pages/Landing.jsx — путь импорта не совпадает.

  Как починить:
  1. Проверь, что файл существует:
       ls src/pages/Landing.jsx
  2. Если расширение .jsx — Vite требует его явно:
       import Landing from './pages/Landing.jsx'
  3. Или добавь resolve.extensions в vite.config.js
```

## Возможности

- 🧠 **Три провайдера на выбор** — Claude, OpenAI или Gemini. Переключение одной командой. Твой ключ, твой счёт, ничего не проксируется.
- 🐚 **Любой shell** — bash, zsh, fish, PowerShell. `wtf init` ставит хук, после этого захват ошибок идёт автоматически.
- ⚡ **Спиннер и красивый Markdown** — ответ рендерится прямо в терминале с цветами, заголовками, выделенными командами.
- 🛡️ **Чистка секретов** — 13 regex-правил перед отправкой в API. Токены, JWT, пароли, email, basic-auth URL, абсолютные пути домашней директории — всё вычищается.
- 💾 **Локальный кеш** — повторные одинаковые ошибки показываются мгновенно из `~/.wtf/cache/`. TTL 30 дней.
- 🔒 **Zero trust** — нет своего бэкенда, нет аккаунтов, нет телеметрии. Только локальный конфиг и исходящие HTTPS на API провайдера.
- 🪶 **Один бинарь** — Go, ~8 МБ, без зависимостей.
- 🌍 **RU + EN** — объяснение на русском по умолчанию, `wtf --lang en` — на английском.

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

## Быстрый старт

```bash
# 1. Настрой провайдера (интерактивный wizard)
wtf config

# 2. Поставь shell-хук, чтобы wtf автоматически читал последнюю ошибку
wtf init

# 3. Перезапусти shell или выполни:
source ~/.zshrc   # или ~/.bashrc / fish config / . $PROFILE

# 4. Запусти любую команду которая упадёт. Затем — wtf.
npm run build     # упало
wtf               # объясни
```

Альтернативы если хук не установлен:

```bash
wtf --rerun                          # перезапустит последнюю команду и поймает stderr
wtf --explain "Error: ECONNREFUSED"  # явный текст
echo "$ERROR_TEXT" | wtf             # пайпом тоже можно (todo)
```

## Команды

```
wtf                        объяснить последнюю ошибку
wtf --rerun                перезапустить последнюю команду и объяснить
wtf --explain "<text>"     объяснить переданный текст
wtf --provider <name>      разово выбрать провайдера (claude|openai|gemini)
wtf --no-cache             не использовать кеш
wtf --lang <ru|en>         язык ответа

wtf init [shell]           установить shell-хук (bash|zsh|fish|powershell)
wtf config                 интерактивная настройка
wtf config show            показать текущий конфиг
wtf config set k=v         задать значение программно
wtf version                версия
```

## Конфигурация

Конфиг лежит в `~/.wtf/config.json` (mode 0600):

```json
{
  "default_provider": "claude",
  "language": "ru",
  "cache_enabled": true,
  "providers": {
    "claude": { "api_key": "sk-ant-...", "model": "claude-haiku-4-5-20251001" },
    "openai": { "api_key": "sk-...",     "model": "gpt-4o-mini" },
    "gemini": { "api_key": "AIza...",    "model": "gemini-2.0-flash" }
  }
}
```

Можно задавать ключи через переменные окружения — это переопределяется конфигом:

- `ANTHROPIC_API_KEY` — для Claude
- `OPENAI_API_KEY` — для OpenAI
- `GEMINI_API_KEY` или `GOOGLE_API_KEY` — для Gemini

## Как это работает

```
┌──────────────┐
│   wtf CLI    │
└──────┬───────┘
       │
       ├──► читает ~/.wtf/last_meta (cmd, exit code) — пишется shell-хуком
       ├──► читает stdout/stderr последней команды (или --rerun)
       │
       ├──► собирает контекст: OS, shell, cwd, git branch, package manager
       │
       ├──► redaction: 13 regex-правил вычищают секреты
       │
       ├──► проверяет локальный кеш по sha256(provider+model+lang+cmd+output)
       │
       ├──► POST в API провайдера (Claude/OpenAI/Gemini)
       │       — Claude использует prompt caching для system-промпта
       │
       └──◄ Markdown-ответ → ANSI-render в терминал
```

### Redaction

Перед отправкой работает следующий список regex-фильтров:

| Класс | Шаблон |
|---|---|
| Anthropic API key | `sk-ant-[A-Za-z0-9_\-]{20,}` |
| OpenAI / generic API key | `sk-[A-Za-z0-9_\-]{20,}` |
| Google API key | `AIza[0-9A-Za-z_\-]{35}` |
| GitHub token | `gh[pousr]_[A-Za-z0-9]{20,}` |
| Slack token | `xox[baprs]-...` |
| JWT | `eyJ...\.eyJ...\....` |
| Bearer token | `Bearer <anything>` |
| AWS access key | `AKIA[0-9A-Z]{16}` |
| AWS secret | `aws_secret*=...` |
| Private keys | `-----BEGIN ... PRIVATE KEY-----...` |
| Generic password/secret/token KV | `password=...`, `token=...`, etc. |
| Basic-auth in URL | `https://user:pass@host` |
| Email | `user@host.tld` |
| Home path | `$HOME → ~` |

При первом запуске показывается консент-баннер с тем, что именно отправляется — после подтверждения он не повторяется.

## Стоимость

Один разбор стандартной ошибки на Claude Haiku 4.5 — порядка $0.0001-0.0003. Это копейки даже при сотне фейлов в день. На gpt-4o-mini и Gemini Flash — сравнимо.

## Сравнение с похожими тулами

| | wtf | tldr | thefuck |
|---|---|---|---|
| Что делает | объясняет конкретную ошибку через AI | показывает man-cheatsheet команды | угадывает опечатку из словаря |
| AI / нейросеть | ✅ Claude / OpenAI / Gemini | ❌ статика | ❌ хардкод |
| Понимает контекст | ✅ stderr, exit, OS, shell, git | ❌ только имя команды | ❌ только командная строка |
| Работает с любой ошибкой | ✅ | ❌ только описанные | ❌ только знакомые правила |
| Защита секретов | ✅ regex-фильтр | n/a | n/a |
| Стоимость | ~копейки за вызов | бесплатно | бесплатно |

Все три могут жить рядом. `tldr` — когда не помнишь синтаксис. `thefuck` — когда опечатался. `wtf` — когда упало и непонятно почему.

## Roadmap

- [ ] `--offline` режим через ollama (локальный LLM)
- [ ] Streaming ответа (сейчас ждём всё одним куском)
- [ ] `wtf --share` — обезличенная ссылка на GitHub Gist для коллег
- [ ] Поддержка `cmd.exe` на Windows
- [ ] `wtf doctor` — диагностика установки хуков
- [ ] Brew tap, scoop, AUR, deb/rpm

## Безопасность

- Конфиг и кеш лежат в `~/.wtf/` с правами `0600` / `0644` (директория).
- Ключи отображаются маскированно (`sk-a…XXXX`).
- Никакой удалённой телеметрии.
- Если найдёшь утечку секрета через regex-фильтр — присылай PR с тестом.

## Вклад

PR welcome. Запустить локально:

```bash
git clone https://github.com/kitay-sudo/wtf
cd wtf
go build -o wtf ./cmd/wtf
./wtf --explain "Error: test"   # для smoke-test
```

Фронтенд лендинга:

```bash
cd frontend
npm install
npm run dev   # http://localhost:5173
```

## Лицензия

MIT © [kitay-sudo](https://github.com/kitay-sudo)
