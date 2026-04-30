package shellhook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	wtfcfg "github.com/bitcoff/wtf/internal/config"
)

type Capture struct {
	Command  string
	ExitCode int
	Output   string
	Source   string
}

func MetaPath() (string, error) {
	dir, err := wtfcfg.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "last_meta"), nil
}

func OutputPath() (string, error) {
	dir, err := wtfcfg.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "last_output"), nil
}

func ReadCapture() (*Capture, error) {
	dir, err := wtfcfg.Dir()
	if err != nil {
		return nil, err
	}

	// Если есть свежий last_meta_capture (от wtfc) — читаем его, потому что
	// last_meta перезатирается каждой следующей prompt-командой (включая wtf).
	cap := &Capture{Source: "hook"}
	captureMeta := filepath.Join(dir, "last_meta_capture")
	regularMeta := filepath.Join(dir, "last_meta")

	chosen := regularMeta
	if st, err := os.Stat(captureMeta); err == nil {
		if rst, rerr := os.Stat(regularMeta); rerr == nil {
			// Берём более свежий из двух.
			if st.ModTime().After(rst.ModTime()) {
				chosen = captureMeta
				cap.Source = "wtfc"
			}
		} else {
			chosen = captureMeta
			cap.Source = "wtfc"
		}
	}

	meta, err := os.ReadFile(chosen)
	if err != nil {
		return nil, fmt.Errorf("no last_meta — установи shell-хук: wtf init (%w)", err)
	}
	for _, line := range strings.Split(string(meta), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch k {
		case "cmd":
			cap.Command = v
		case "exit":
			cap.ExitCode, _ = strconv.Atoi(strings.TrimSpace(v))
		}
	}

	outP, _ := OutputPath()
	if data, err := os.ReadFile(outP); err == nil {
		cap.Output = stripANSI(string(data))
	}
	if cap.Output == "" && cap.Command == "" {
		return nil, fmt.Errorf("last_meta пустой — выполни команду в shell с установленным хуком")
	}
	return cap, nil
}

// stripANSI убирает escape-последовательности (цвета, перемещения курсора)
// из захваченного через `tee` вывода. AI работает с чистым текстом точнее
// и redaction-фильтр не путается.
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\x07]*\x07|\x1b[=>]|\x1b\([AB012]`)

func stripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

func Rerun(command string) (*Capture, error) {
	if command == "" {
		return nil, fmt.Errorf("пустая команда для rerun")
	}
	shell := detectCurrentShell()
	if shell == "" {
		return nil, fmt.Errorf("не удалось определить shell")
	}
	args := []string{"-c", command}
	if shell == "powershell" || shell == "pwsh" {
		args = []string{"-NoProfile", "-Command", command}
	}
	cmd := exec.Command(shell, args...)
	cmd.Env = append(os.Environ(), "WTF_RERUN=1")
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return &Capture{
		Command:  command,
		ExitCode: exitCode,
		Output:   string(out),
		Source:   "rerun",
	}, nil
}

func detectCurrentShell() string {
	if v := os.Getenv("WTF_SHELL"); v != "" {
		return v
	}
	if v := os.Getenv("SHELL"); v != "" {
		return v
	}
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("pwsh"); err == nil {
			return "pwsh"
		}
		return "powershell"
	}
	return "/bin/sh"
}

// Bash-хук: пишет в last_meta команду и exit-код после каждой команды,
// плюс предоставляет функцию-обёртку '?>' для явного захвата stdout/stderr.
//
// Идеология: захватить вывод произвольной команды через DEBUG trap+tee
// в bash оказалось ненадёжно (DEBUG срабатывает на подкоманды,
// race с асинхронным tee, ломаются TUI). Поэтому даём явный
// контроль юзеру:
//   $ ?> service nginx status   # обёртка пишет stdout+stderr в last_output
//   $ wtf                        # читает захваченный вывод
//
// Альтернатива без обёртки — pipe: 'service nginx status 2>&1 | wtf'.
const bashHook = `# wtf shell hook (bash)
__wtf_dir="$HOME/.wtf"
__wtf_capture="$__wtf_dir/last_output"
__wtf_meta="$__wtf_dir/last_meta"
__wtf_cmd=""
mkdir -p "$__wtf_dir"

__wtf_preexec() {
  __wtf_cmd="$BASH_COMMAND"
  printf 'cmd=%s\nexit=0\nts=%s\n' "$__wtf_cmd" "$(date +%s)" > "$__wtf_meta"
}
trap '__wtf_preexec' DEBUG

__wtf_precmd() {
  local ec=$?
  if [ -n "$__wtf_cmd" ]; then
    printf 'cmd=%s\nexit=%s\nts=%s\n' "$__wtf_cmd" "$ec" "$(date +%s)" > "$__wtf_meta"
  fi
}
case ";${PROMPT_COMMAND};" in
  *";__wtf_precmd;"*) ;;
  *) PROMPT_COMMAND="__wtf_precmd;${PROMPT_COMMAND}" ;;
esac

# Префикс-обёртка: 'wtfc <команда>' выполняет команду и захватывает её
# stdout+stderr в ~/.wtf/last_output, чтобы потом 'wtf' их прочитал.
# Использует 'tee', не ломает интерактивные TUI.
wtfc() {
  if [ "$#" -eq 0 ]; then
    printf 'usage: wtfc <команда>\n  захватывает stdout+stderr команды в ~/.wtf/last_output,\n  затем напиши: wtf\n' >&2
    return 2
  fi
  local cmd_str="$*"
  "$@" 2>&1 | tee "$__wtf_capture"
  local ec=${PIPESTATUS[0]}
  # Пишем в last_meta_capture чтобы prompt-hook не перезатёр после wtfc.
  printf 'cmd=%s\nexit=%s\nts=%s\n' "$cmd_str" "$ec" "$(date +%s)" > "$__wtf_dir/last_meta_capture"
  return "$ec"
}
`

const zshHook = `# wtf shell hook (zsh)
__wtf_dir="$HOME/.wtf"
__wtf_capture="$__wtf_dir/last_output"
__wtf_meta="$__wtf_dir/last_meta"
mkdir -p "$__wtf_dir"

__wtf_preexec() {
  print -r -- "cmd=$1" > "$__wtf_meta"
  print -r -- "exit=0" >> "$__wtf_meta"
  print -r -- "ts=$(date +%s)" >> "$__wtf_meta"
}
__wtf_precmd() {
  local ec=$?
  local cmd
  cmd="$(fc -ln -1 2>/dev/null | sed 's/^[[:space:]]*//')"
  print -r -- "cmd=$cmd" > "$__wtf_meta"
  print -r -- "exit=$ec" >> "$__wtf_meta"
  print -r -- "ts=$(date +%s)" >> "$__wtf_meta"
}
autoload -Uz add-zsh-hook
add-zsh-hook preexec __wtf_preexec
add-zsh-hook precmd __wtf_precmd

# Префикс-обёртка: 'wtfc <команда>' захватывает вывод в ~/.wtf/last_output.
wtfc() {
  if [ "$#" -eq 0 ]; then
    print -r -- "usage: wtfc <команда>" >&2
    return 2
  fi
  local cmd_str="$*"
  "$@" 2>&1 | tee "$__wtf_capture"
  local ec=${pipestatus[1]}
  print -r -- "cmd=$cmd_str" > "$__wtf_dir/last_meta_capture"
  print -r -- "exit=$ec" >> "$__wtf_dir/last_meta_capture"
  print -r -- "ts=$(date +%s)" >> "$__wtf_dir/last_meta_capture"
  return "$ec"
}
`

const fishHook = `# wtf shell hook (fish)
set -g __wtf_dir "$HOME/.wtf"
set -g __wtf_meta "$__wtf_dir/last_meta"
set -g __wtf_capture "$__wtf_dir/last_output"
mkdir -p $__wtf_dir

function __wtf_preexec --on-event fish_preexec
    printf 'cmd=%s\nexit=0\nts=%s\n' "$argv" (date +%s) > $__wtf_meta
end

function __wtf_postexec --on-event fish_postexec
    set -l ec $status
    printf 'cmd=%s\nexit=%s\nts=%s\n' "$argv" $ec (date +%s) > $__wtf_meta
end

# Префикс-обёртка для fish.
function wtfc
    if test (count $argv) -eq 0
        echo "usage: wtfc <команда>" >&2
        return 2
    end
    set -l cmd_str (string join " " $argv)
    $argv 2>&1 | tee $__wtf_capture
    set -l ec $pipestatus[1]
    printf 'cmd=%s\nexit=%s\nts=%s\n' "$cmd_str" $ec (date +%s) > $__wtf_dir/last_meta_capture
    return $ec
end
`

const pwshHook = `# wtf shell hook (PowerShell)
$__wtfDir = Join-Path $HOME ".wtf"
$__wtfMeta = Join-Path $__wtfDir "last_meta"
if (-not (Test-Path $__wtfDir)) { New-Item -ItemType Directory -Path $__wtfDir | Out-Null }

$global:__wtfPromptOriginal = $function:prompt
function global:prompt {
    $ec = if ($?) { 0 } else { 1 }
    $lastCmd = (Get-History -Count 1).CommandLine
    if ($lastCmd) {
        $ts = [int][double]::Parse((Get-Date -UFormat %s))
        @("cmd=$lastCmd","exit=$ec","ts=$ts") | Set-Content -Path $__wtfMeta -Encoding utf8
    }
    & $global:__wtfPromptOriginal
}
`

type Shell string

const (
	ShellBash       Shell = "bash"
	ShellZsh        Shell = "zsh"
	ShellFish       Shell = "fish"
	ShellPowerShell Shell = "powershell"
)

func DetectInstall() Shell {
	if runtime.GOOS == "windows" {
		return ShellPowerShell
	}
	if v := os.Getenv("SHELL"); v != "" {
		base := filepath.Base(v)
		switch base {
		case "zsh":
			return ShellZsh
		case "fish":
			return ShellFish
		case "bash", "sh":
			return ShellBash
		}
	}
	return ShellBash
}

func HookScript(s Shell) string {
	switch s {
	case ShellBash:
		return bashHook
	case ShellZsh:
		return zshHook
	case ShellFish:
		return fishHook
	case ShellPowerShell:
		return pwshHook
	}
	return ""
}

func RcPath(s Shell) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch s {
	case ShellBash:
		return filepath.Join(home, ".bashrc"), nil
	case ShellZsh:
		return filepath.Join(home, ".zshrc"), nil
	case ShellFish:
		return filepath.Join(home, ".config", "fish", "conf.d", "wtf.fish"), nil
	case ShellPowerShell:
		return powershellProfilePath()
	}
	return "", fmt.Errorf("unknown shell: %s", s)
}

func powershellProfilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), nil
}

const marker = "# >>> wtf shell hook >>>"
const markerEnd = "# <<< wtf shell hook <<<"

func Install(s Shell) (string, error) {
	rc, err := RcPath(s)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(rc), 0o755); err != nil {
		return "", err
	}
	hookFile := rc + ".wtf-hook"

	if err := os.WriteFile(hookFile, []byte(HookScript(s)), 0o644); err != nil {
		return "", err
	}

	srcLine := sourceLine(s, hookFile)
	existing, _ := os.ReadFile(rc)
	if strings.Contains(string(existing), marker) {
		return rc, nil
	}
	block := fmt.Sprintf("\n%s (installed %s)\n%s\n%s\n",
		marker, time.Now().Format("2006-01-02"), srcLine, markerEnd)
	f, err := os.OpenFile(rc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.WriteString(block); err != nil {
		return "", err
	}
	return rc, nil
}

func sourceLine(s Shell, hookFile string) string {
	switch s {
	case ShellPowerShell:
		return fmt.Sprintf(". '%s'", hookFile)
	case ShellFish:
		return fmt.Sprintf("source '%s'", hookFile)
	default:
		return fmt.Sprintf("[ -f \"%s\" ] && . \"%s\"", hookFile, hookFile)
	}
}
