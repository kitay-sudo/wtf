# wtf — AI-объяснялка для терминала

> Видишь в терминале что-то непонятное — упавшую команду, статус сервиса, дамп конфига, странный вывод утилиты — пиши `wtf`. Получаешь короткое объяснение и при необходимости 2-3 готовых действия. Один Go-бинарь, три AI-провайдера на выбор, MIT.

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

`wtf` — это CLI-утилита, которая берёт **последний вывод твоего терминала** (stdout + stderr последней команды) и объясняет его человеческими словами. Это не только про ошибки: статус сервиса, дамп конфига, JSON от API, лог-файл, вывод незнакомой утилиты — что угодно, что у тебя сейчас в окне терминала. Если есть что чинить — добавит 2-3 готовых действия.

### Пример 1 — упавшая команда

```
$ npm run build
Error: Cannot find module './pages/Landing' imported from src/App.jsx

$ wtf
  ⠹  анализ через claude (haiku-4-5)...
  ✓  Готово · 1.2s · claude

  Что случилось: Vite не находит файл src/pages/Landing.jsx — путь импорта не совпадает.

  Как починить:
  1. Проверь, что файл существует:
       ls src/pages/Landing.jsx
  2. Если расширение .jsx — Vite требует его явно:
       import Landing from './pages/Landing.jsx'
  3. Или добавь resolve.extensions в vite.config.js
```

### Пример 2 — просто разобраться в выводе

Команда отработала успешно, но вывод нечитаемый — `wtf` это тоже умеет.

```
$ systemctl status nginx
● nginx.service - A high performance web server and a reverse proxy server
     Loaded: loaded (/lib/systemd/system/nginx.service; enabled; vendor preset: enabled)
     Active: active (running) since Tue 2026-04-29 14:22:11 UTC; 18h ago
       Docs: man:nginx(8)
    Process: 1247 ExecStartPre=/usr/sbin/nginx -t -q -g daemon on; master_process on; (code=exited, status=0/SUCCESS)
    Process: 1248 ExecStart=/usr/sbin/nginx -g daemon on; master_process on; (code=exited, status=0/SUCCESS)
   Main PID: 1249 (nginx)
      Tasks: 3 (limit: 4915)
     Memory: 12.4M
        CPU: 142ms
     CGroup: /system.slice/nginx.service
             ├─1249 "nginx: master process /usr/sbin/nginx -g daemon on; master_process on;"
             ├─1250 "nginx: worker process"
             └─1251 "nginx: worker process"

$ wtf
  ✓  Готово · 0.9s · claude

  Что это: nginx работает нормально — сервис запущен 18 часов назад,
  master + 2 worker процесса, потребляет 12 МБ памяти.

  На что обратить внимание:
  • Active: active (running) — всё ок
  • ExecStartPre с status=0/SUCCESS — конфиг прошёл проверку при старте
  • 2 worker process — стандартно для systemd-юнита из коробки

  Если что-то нужно: nginx -t (проверить конфиг), journalctl -u nginx -f (живые логи).
```

## Возможности

- 🧠 **Три провайдера на выбор** — Claude, OpenAI или Gemini. Переключение одной командой. Твой ключ, твой счёт, ничего не проксируется.
- 🐚 **Любой shell** — bash, zsh, fish, PowerShell. `wtf init` ставит хук, после этого захват ошибок идёт автоматически.
- ⚡ **Спиннер и красивый Markdown** — ответ рендерится прямо в терминале с цветами, заголовками, выделенными командами.
- 🛡️ **Чистка секретов** — 13 regex-правил перед отправкой в API. Токены, JWT, пароли, email, basic-auth URL, абсолютные пути домашней директории — всё вычищается.
- 💾 **Локальный кеш** — повторные одинаковые запросы возвращаются мгновенно из `~/.wtf/cache/`. TTL 30 дней.
- 🔒 **Zero trust** — нет своего бэкенда, нет аккаунтов, нет телеметрии. Только локальный конфиг и исходящие HTTPS на API провайдера.
- 🪶 **Один бинарь** — Go, ~8 МБ, без зависимостей.
- 🌍 **RU + EN** — объяснение на русском по умолчанию, `wtf --lang en` — на английском.

## Быстрый старт

```bash
# 1. Настрой провайдера (интерактивный wizard)
wtf config

# 2. Поставь shell-хук, чтобы wtf автоматически читал последний вывод терминала
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
wtf                        объяснить последний вывод в терминале
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

Если проект помог — поставь ⭐ на [github.com/kitay-sudo/wtf](https://github.com/kitay-sudo/wtf). Это бесплатно для тебя и реально помогает: чем больше звёзд, тем проще другим находить инструмент.

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

### Поддержать криптой

Если хочется отблагодарить напрямую — буду рад. После доната напиши в Telegram [@kitay9](https://t.me/kitay9) свой ник, добавлю в стену чести на лендинге.

| Сеть | Адрес |
|---|---|
| **USDT** · TRON (TRC20) | `TF9F2FPkreHVfbe8tZtn4V76j3jLo4SeXM` |
| **TON** · The Open Network | `UQBl88kXWJWyHkDPkWNYQwwSCiCAIfA2DiExtZElwJFlIc1o` |

### Рекомендую: Timeweb Cloud

[Timeweb Cloud](https://timeweb.cloud/?i=104289) — российский VPS-хостинг, на котором живут наши боевые сервера. Быстрая панель, NVMe, развёртывание за минуту. Если возьмёшь сервер — могу помочь с первичной настройкой: пиши в Telegram [@kitay9](https://t.me/kitay9).

*Ссылка партнёрская (`ad`).*

### Выпуск нового релиза

Релиз — это git-тег `vX.Y.Z`, всё остальное автоматизировано:

```bash
# Windows
scripts\release.bat patch    # bug-fix:    v0.1.0 → v0.1.1
scripts\release.bat minor    # фича:        v0.1.0 → v0.2.0
scripts\release.bat major    # breaking:    v0.1.0 → v1.0.0
scripts\release.bat v0.5.0   # явная версия

# macOS / Linux / Git Bash
./scripts/release.sh patch
```

Скрипт:

1. Проверяет что ты на `main`, working tree чистый и синхронизирован с `origin`.
2. Считает следующую версию по последнему тегу (или берёт явную).
3. Показывает список коммитов с прошлого релиза и просит подтверждения.
4. Создаёт annotated-тег и пушит его в `origin`.

После пуша тега запускается [`.github/workflows/release.yml`](.github/workflows/release.yml):

- matrix-сборка под 6 платформ (`linux/darwin/windows × amd64/arm64`) с прокинутой версией через `-ldflags "-X main.version=$TAG"`
- архивы `wtf_<os>_<arch>.tar.gz` (или `.zip` для Windows) с README + LICENSE внутри
- `SHA256SUMS.txt` для верификации
- GitHub Release с авто-сгенерированными notes (по коммитам с прошлого тега)

Через 2-3 минуты бинари висят на странице Releases, а `install.sh` / `install.ps1` начинают видеть новую версию через GitHub API.

## Лицензия

MIT © [kitay-sudo](https://github.com/kitay-sudo)
