//go:build !windows

package exec

import (
	"os/exec"
	"syscall"
)

// setProcessGroup помещает процесс в новую process group, чтобы
// killProcessGroup мог убить всё дерево разом (через -pgid).
func setProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// killProcessGroup посылает SIGKILL всей process group.
// Используется при таймауте, когда дочерние процессы (pm2 logs, journalctl -f
// и т.п.) держат stdout pipe и cmd.Wait() висит.
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		_ = cmd.Process.Kill()
		return
	}
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
}
