package shellhook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	metaP, err := MetaPath()
	if err != nil {
		return nil, err
	}
	meta, err := os.ReadFile(metaP)
	if err != nil {
		return nil, fmt.Errorf("no last_meta — установи shell-хук: wtf init (%w)", err)
	}
	cap := &Capture{Source: "hook"}
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
		cap.Output = string(data)
	}
	if cap.Output == "" && cap.Command == "" {
		return nil, fmt.Errorf("last_meta пустой — выполни команду в shell с установленным хуком")
	}
	return cap, nil
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
`

const zshHook = `# wtf shell hook (zsh)
__wtf_dir="$HOME/.wtf"
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
`

const fishHook = `# wtf shell hook (fish)
set -g __wtf_dir "$HOME/.wtf"
set -g __wtf_meta "$__wtf_dir/last_meta"
mkdir -p $__wtf_dir

function __wtf_preexec --on-event fish_preexec
    printf 'cmd=%s\nexit=0\nts=%s\n' "$argv" (date +%s) > $__wtf_meta
end

function __wtf_postexec --on-event fish_postexec
    set -l ec $status
    printf 'cmd=%s\nexit=%s\nts=%s\n' "$argv" $ec (date +%s) > $__wtf_meta
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
