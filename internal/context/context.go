package context

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Info struct {
	OS         string
	Shell      string
	Cwd        string
	GitBranch  string
	PkgManager string
}

func Collect() Info {
	cwd, _ := os.Getwd()
	return Info{
		OS:         osString(),
		Shell:      detectShell(),
		Cwd:        cwd,
		GitBranch:  gitBranch(cwd),
		PkgManager: detectPkgManager(cwd),
	}
}

func osString() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

func detectShell() string {
	if v := os.Getenv("WTF_SHELL"); v != "" {
		return v
	}
	if v := os.Getenv("SHELL"); v != "" {
		return filepath.Base(v)
	}
	if os.Getenv("PSModulePath") != "" {
		return "powershell"
	}
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return ""
}

func gitBranch(cwd string) string {
	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func detectPkgManager(cwd string) string {
	checks := []struct {
		file string
		name string
	}{
		{"package-lock.json", "npm"},
		{"yarn.lock", "yarn"},
		{"pnpm-lock.yaml", "pnpm"},
		{"bun.lockb", "bun"},
		{"Cargo.toml", "cargo"},
		{"go.mod", "go"},
		{"pyproject.toml", "python"},
		{"requirements.txt", "pip"},
		{"Gemfile", "bundler"},
		{"composer.json", "composer"},
		{"pom.xml", "maven"},
		{"build.gradle", "gradle"},
		{"build.gradle.kts", "gradle"},
	}
	dir := cwd
	for i := 0; i < 4; i++ {
		for _, c := range checks {
			if _, err := os.Stat(filepath.Join(dir, c.file)); err == nil {
				return c.name
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
