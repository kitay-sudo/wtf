# Changelog

Все значимые изменения в проекте описываются здесь. Формат — [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/), версионирование — [SemVer](https://semver.org/lang/ru/).

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
