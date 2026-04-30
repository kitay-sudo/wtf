#!/usr/bin/env bash
# wtf — one-command installer / updater / uninstaller.
#
# Usage:
#   # Первая установка ИЛИ обновление (одна команда):
#   curl -sSL https://raw.githubusercontent.com/kitay-sudo/wtf/main/install.sh | sudo bash
#
#   # Принудительная переустановка (снесёт ~/.wtf конфиг!):
#   curl -sSL .../install.sh | sudo bash -s -- --reinstall
#
#   # Только удаление:
#   curl -sSL .../install.sh | sudo bash -s -- --uninstall
#
# Закрепить версию: WTF_VERSION=v0.1.0 перед командой.

set -euo pipefail

REPO="kitay-sudo/wtf"
INSTALL_DIR="/usr/local/bin"
BIN_NAME="wtf"

MODE="auto"
# Под `set -u` пустой "$@" в for-loop может ругнуться на старых bash —
# раскрываем явно через ${@:-}, плюс добавляем условие.
if [[ $# -gt 0 ]]; then
  for arg in "$@"; do
    case "$arg" in
      --reinstall) MODE="reinstall" ;;
      --uninstall) MODE="uninstall" ;;
      -h|--help)
        sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
        exit 0
        ;;
      *) echo "Неизвестный флаг: $arg" >&2; exit 1 ;;
    esac
  done
fi

# ---------- coloring ----------
if [[ -t 1 ]]; then
  C_OK=$'\e[33m'; C_WARN=$'\e[33m'; C_ERR=$'\e[31m'; C_DIM=$'\e[2m'
  C_BOLD=$'\e[1m'; C_CYAN=$'\e[36m'; C_RESET=$'\e[0m'
else
  C_OK=""; C_WARN=""; C_ERR=""; C_DIM=""; C_BOLD=""; C_CYAN=""; C_RESET=""
fi

START_TS=$SECONDS
_ts() {
  local elapsed=$((SECONDS - START_TS))
  printf "%02d:%02d" $((elapsed / 60)) $((elapsed % 60))
}

step() { printf "  %s%s%s  %s→%s  %s\n" "$C_DIM" "$(_ts)" "$C_RESET" "$C_CYAN" "$C_RESET" "$*"; }
ok()   { printf "  %s%s%s  %s✓%s  %s\n" "$C_DIM" "$(_ts)" "$C_RESET" "$C_OK" "$C_RESET" "$*"; }
info() { printf "  %s%s%s  %sⓘ%s  %s\n" "$C_DIM" "$(_ts)" "$C_RESET" "$C_DIM" "$C_RESET" "$*"; }
warn() { printf "  %s%s%s  %s⚠%s  %s\n" "$C_DIM" "$(_ts)" "$C_RESET" "$C_WARN" "$C_RESET" "$*" >&2; }
err()  { printf "  %s%s%s  %s✗%s  %s\n" "$C_DIM" "$(_ts)" "$C_RESET" "$C_ERR" "$C_RESET" "$*" >&2; }
die()  { err "$*"; exit 1; }

header() {
  local version="$1" os="$2" arch="$3"
  printf "\n  %s[!?]  wtf installer%s · %s%s%s · %s%s/%s%s\n\n" \
    "$C_BOLD" "$C_RESET" "$C_OK" "$version" "$C_RESET" "$C_CYAN" "$os" "$arch" "$C_RESET"
}

# ---------- detection ----------
detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       die "Не поддерживаемая ОС: $(uname -s). Для Windows используй install.ps1" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) die "Не поддерживаемая архитектура: $(uname -m)" ;;
  esac
}

latest_version() {
  local response
  response=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest")

  # No releases yet → API returns {"message":"Not Found"} with status 404.
  # Detect this and explain instead of failing silently.
  if echo "$response" | grep -q '"message": *"Not Found"'; then
    err "У репозитория ${REPO} ещё нет ни одного релиза."
    err ""
    err "Маинтейнеру: выпусти первый релиз —"
    err "  scripts\\release.bat v0.1.0   (Windows)"
    err "  ./scripts/release.sh v0.1.0   (Linux/macOS)"
    err ""
    err "После того как workflow release.yml завершится, эта команда заработает."
    err ""
    err "Альтернатива — собрать из исходников:"
    err "  git clone https://github.com/${REPO} && cd wtf && go build -o wtf ./cmd/wtf"
    exit 1
  fi

  echo "$response" \
    | grep -E '"tag_name"' \
    | head -1 \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

# ---------- main flows ----------
do_uninstall() {
  step "удаляю бинарник из ${INSTALL_DIR}/${BIN_NAME}"
  rm -f "${INSTALL_DIR}/${BIN_NAME}"
  ok "wtf удалён"
  info "конфиг и кеш в ~/.wtf/ оставлены — удали вручную если нужно: rm -rf ~/.wtf"
}

TMP_DIR=""
cleanup() {
  # `set -u` + EXIT trap = переменная должна быть инициализирована, иначе
  # скрипт падает после успешной установки с "tmp: unbound variable".
  if [[ -n "${TMP_DIR:-}" && -d "${TMP_DIR}" ]]; then
    rm -rf "${TMP_DIR}"
  fi
}
trap cleanup EXIT

do_install() {
  local os arch version url
  os=$(detect_os)
  arch=$(detect_arch)
  version="${WTF_VERSION:-$(latest_version)}"
  [[ -z "$version" ]] && die "не удалось получить latest version с GitHub"

  header "$version" "$os" "$arch"

  url="https://github.com/${REPO}/releases/download/${version}/wtf_${os}_${arch}.tar.gz"
  TMP_DIR=$(mktemp -d)

  step "скачиваю $url"
  if ! curl -sSL -o "$TMP_DIR/wtf.tar.gz" "$url"; then
    die "не удалось скачать релиз — проверь, что версия $version существует"
  fi
  ok "скачано"

  step "распаковываю"
  tar -xzf "$TMP_DIR/wtf.tar.gz" -C "$TMP_DIR"
  [[ -f "$TMP_DIR/wtf" ]] || die "в архиве нет бинарника wtf"
  ok "распаковано"

  step "устанавливаю в ${INSTALL_DIR}/${BIN_NAME}"
  install -m 0755 "$TMP_DIR/wtf" "${INSTALL_DIR}/${BIN_NAME}"
  ok "установлено: $(${INSTALL_DIR}/${BIN_NAME} version)"

  printf "\n"
  printf "  %s[!?] готово!%s\n\n" "$C_BOLD" "$C_RESET"
  printf "  Дальше:\n"
  printf "    %s$%s wtf config       %s# настроить провайдера и ключ%s\n" "$C_DIM" "$C_RESET" "$C_DIM" "$C_RESET"
  printf "    %s$%s wtf init         %s# поставить shell-хук%s\n" "$C_DIM" "$C_RESET" "$C_DIM" "$C_RESET"
  printf "    %s$%s wtf              %s# объяснить последнюю ошибку%s\n\n" "$C_DIM" "$C_RESET" "$C_DIM" "$C_RESET"
}

case "$MODE" in
  uninstall) do_uninstall ;;
  reinstall)
    do_uninstall
    rm -rf "$HOME/.wtf"
    info "конфиг ~/.wtf удалён"
    do_install
    ;;
  auto) do_install ;;
esac
