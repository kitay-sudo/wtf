//go:build windows

package exec

import "os/exec"

// На Windows process groups работают иначе (Job Objects), и таймаут-сценарий
// у нас в основном для Unix. Используем простой Kill() — этого достаточно
// для большинства команд.
func setProcessGroup(cmd *exec.Cmd) {
	// no-op
}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
}
