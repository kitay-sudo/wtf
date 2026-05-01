// Package exec классифицирует команды на безопасные (запускаем сами)
// и destructive (показываем юзеру, не запускаем) и выполняет первые.
//
// Идеология: лучше отказать выполнить странную команду и спросить юзера,
// чем случайно стереть прод. Поэтому:
//   - Whitelist первого токена → безопасно.
//   - Blacklist первого токена ИЛИ опасные паттерны (sudo, rm, >, mkfs, dd) → destructive.
//   - Всё остальное → unknown, спрашиваем юзера.
package exec

import (
	"regexp"
	"strings"
)

type Class int

const (
	ClassSafe Class = iota
	ClassDestructive
	ClassUnknown
)

func (c Class) String() string {
	switch c {
	case ClassSafe:
		return "safe"
	case ClassDestructive:
		return "destructive"
	}
	return "unknown"
}

// Read-only утилиты, безопасные для авто-запуска. Покрывают типовую диагностику:
// процесс, сеть, диски, пакеты, логи, git, docker (только status/logs/inspect).
// Ничего из этого не модифицирует систему даже косвенно.
var safeFirstTokens = map[string]bool{
	// фс / навигация
	"ls": true, "ll": true, "pwd": true, "stat": true, "file": true,
	"realpath": true, "readlink": true, "tree": true, "find": true,
	"locate": true, "which": true, "whereis": true, "type": true,
	// контент
	"cat": true, "tac": true, "head": true, "tail": true, "less": true,
	"more": true, "wc": true, "sort": true, "uniq": true, "tr": true,
	"cut": true, "awk": true, "sed": true, "grep": true, "egrep": true,
	"fgrep": true, "rg": true, "ag": true, "diff": true, "cmp": true,
	"hexdump": true, "xxd": true, "od": true, "strings": true, "jq": true,
	"yq": true, "column": true, "fold": true, "nl": true,
	// процессы / система
	"ps": true, "top": true, "htop": true, "pidof": true, "pgrep": true,
	"uptime": true, "uname": true, "hostname": true, "id": true,
	"whoami": true, "groups": true, "users": true, "who": true, "w": true,
	"last": true, "lastlog": true, "env": true, "printenv": true,
	"date": true, "cal": true, "tty": true, "free": true, "vmstat": true,
	"iostat": true, "mpstat": true, "lscpu": true, "lsblk": true,
	"lspci": true, "lsusb": true, "lshw": true, "dmidecode": true,
	"sensors": true,
	// диски
	"df": true, "du": true, "mount": true, "findmnt": true, "blkid": true,
	"smartctl": true, "fdisk": true, "parted": true,
	// сеть
	"ip": true, "ifconfig": true, "ss": true, "netstat": true, "lsof": true,
	"ping": true, "traceroute": true, "tracepath": true, "mtr": true,
	"dig": true, "host": true, "nslookup": true, "whois": true,
	"arp": true, "route": true, "nmap": true,
	// пакеты (только query-операции, см. также blacklist на install/remove)
	"dpkg": true, "rpm": true, "snap": true, "flatpak": true,
	"brew": true, "pacman": true, "yum": true, "dnf": true, "apt": true,
	"apt-cache": true,
	// systemd / journal — read-only режимы. Опасные глаголы (start/stop/restart)
	// перехватываются isDestructive() ниже.
	"systemctl": true, "journalctl": true, "service": true,
	"loginctl": true, "timedatectl": true, "hostnamectl": true,
	// docker — только safe-подкоманды, остальное в isDestructive
	"docker": true, "podman": true, "kubectl": true, "helm": true,
	// git — read-only глаголы, push/reset --hard в isDestructive
	"git": true,
	// разное
	"echo": true, "printf": true, "true": true, "false": true,
	"nginx": true, "apachectl": true, "httpd": true,
	"php": true, "python": true, "python3": true, "node": true,
	"go": true, "ruby": true, "perl": true,
	"crontab": true, "iptables": true, "nft": true, "ufw": true,
	"ssh-add": true,
	"openssl": true, "certbot": true,
	// process managers — read-only подкоманды разрешены, write-операции
	// (start/stop/restart/delete/save/...) перехватываются isDestructiveSubcommand.
	"pm2": true, "supervisorctl": true, "forever": true,
}

// Команды которые модифицируют систему — никогда не запускаем сами.
var destructiveFirstTokens = map[string]bool{
	"rm": true, "rmdir": true, "unlink": true, "shred": true,
	"mv": true, "dd": true, "mkfs": true, "fsck": true, "wipefs": true,
	"chmod": true, "chown": true, "chgrp": true, "chattr": true,
	"setfacl": true,
	"kill": true, "killall": true, "pkill": true, "skill": true,
	"reboot": true, "shutdown": true, "halt": true, "poweroff": true,
	"useradd": true, "userdel": true, "usermod": true, "passwd": true,
	"groupadd": true, "groupdel": true,
	"sudo": true, "su": true, "doas": true,
	"pip": true, "pip3": true, "npm": true, "yarn": true, "pnpm": true,
	"gem": true, "cargo": true, "cabal": true, "stack": true,
	"curl": true, // curl потенциально безопасен (-I, --head), но `curl … | sh` опасен — проще занести в destructive
	"wget": true,
	"tar": true, "unzip": true, "zip": true, "gzip": true, "gunzip": true,
	// tar/zip формально могут быть безопасны (-t, -l), но `tar -xf` извлекает.
	// Проще всё семейство в destructive — пусть юзер сам решит.
}

// isDestructiveSubcommand — для команд из safeFirstTokens, у которых второй
// токен меняет всё. systemctl restart, git push, docker rm, apt install и т.д.
func isDestructiveSubcommand(tokens []string) bool {
	if len(tokens) < 2 {
		return false
	}
	cmd := tokens[0]
	sub := tokens[1]

	switch cmd {
	case "systemctl", "service":
		switch sub {
		case "start", "stop", "restart", "reload", "enable", "disable",
			"mask", "unmask", "kill", "reset-failed", "edit", "set-property",
			"daemon-reload", "isolate", "rescue", "emergency":
			return true
		}
	case "git":
		switch sub {
		case "push", "reset", "clean", "rebase", "merge", "checkout",
			"switch", "restore", "rm", "mv", "commit", "add",
			"cherry-pick", "revert", "stash", "tag", "branch":
			// branch -d / tag -d тоже опасны; всё это семейство — destructive
			return true
		}
	case "docker", "podman":
		switch sub {
		case "run", "rm", "rmi", "start", "stop", "restart", "kill",
			"pause", "unpause", "create", "build", "push", "pull",
			"exec", "commit", "cp", "load", "save", "import", "export",
			"network", "volume", "system", "compose", "swarm":
			return true
		}
	case "kubectl":
		switch sub {
		case "apply", "create", "delete", "edit", "patch", "replace",
			"scale", "rollout", "drain", "cordon", "uncordon", "taint",
			"label", "annotate", "exec", "port-forward":
			return true
		}
	case "helm":
		switch sub {
		case "install", "upgrade", "uninstall", "delete", "rollback",
			"dependency", "package", "push", "repo":
			return true
		}
	case "apt", "apt-get":
		switch sub {
		case "install", "remove", "purge", "autoremove", "upgrade",
			"dist-upgrade", "full-upgrade", "update", "clean", "autoclean",
			"reinstall", "edit-sources":
			return true
		}
	case "dpkg":
		switch sub {
		case "-i", "--install", "-r", "--remove", "-P", "--purge",
			"--configure", "--unpack":
			return true
		}
	case "yum", "dnf":
		switch sub {
		case "install", "remove", "erase", "update", "upgrade",
			"reinstall", "downgrade", "autoremove", "clean":
			return true
		}
	case "pacman":
		// pacman -S, -R, -U, -Sy, -Syu и т.д. — всё destructive
		if strings.HasPrefix(sub, "-S") || strings.HasPrefix(sub, "-R") || strings.HasPrefix(sub, "-U") {
			return true
		}
	case "snap", "flatpak":
		switch sub {
		case "install", "remove", "uninstall", "refresh", "update":
			return true
		}
	case "brew":
		switch sub {
		case "install", "uninstall", "remove", "upgrade", "reinstall",
			"link", "unlink", "tap", "untap":
			return true
		}
	case "iptables", "nft", "ufw":
		// iptables -A, -D, -F, -I, -P, -R, -X, -Z; nft add/delete/flush; ufw allow/deny/...
		// Всё что не показ правил → destructive. Проще опираться на исключения:
		switch sub {
		case "-L", "-S", "-V", "--list", "list", "list-rules", "show",
			"status", "version":
			return false
		}
		return true
	case "crontab":
		// crontab -l — read; crontab -e/-r — write. crontab без аргументов
		// читает stdin (потенциально destructive). Считаем destructive по умолчанию.
		if sub == "-l" || sub == "--list" {
			return false
		}
		return true
	case "openssl":
		// openssl genrsa / req / x509 -req — генерация ключей/CSR/сертов.
		// Обычно безопасно (создаёт файлы), но в контексте sysadmin-агента
		// лучше показать юзеру, чтобы он понимал что генерится.
		switch sub {
		case "genrsa", "genpkey", "req", "ca", "rsa", "ec":
			return true
		}
	case "certbot":
		// certbot certificates — read; certbot certonly/renew/revoke — write.
		switch sub {
		case "certificates", "show_account", "plugins":
			return false
		}
		return true
	case "go":
		switch sub {
		case "install", "get", "mod", "clean", "build", "run":
			return true
		}
	case "pm2":
		switch sub {
		case "start", "stop", "restart", "reload", "delete", "kill",
			"save", "resurrect", "update", "install", "uninstall",
			"flush", "reloadLogs", "startup", "unstartup", "scale",
			"reset", "monit":
			return true
		}
	case "supervisorctl":
		switch sub {
		case "start", "stop", "restart", "reload", "update", "shutdown",
			"add", "remove", "clear":
			return true
		}
	case "forever":
		switch sub {
		case "start", "stop", "stopall", "restart", "restartall":
			return true
		}
	}
	return false
}

// Опасные паттерны, не привязанные к первому токену.
var (
	rePipeShell      = regexp.MustCompile(`\|\s*(sh|bash|zsh|fish|sudo)\b`)
	reRedirectWrite  = regexp.MustCompile(`(?:^|[^>])>\s*[^>]`) // > и >> без двойного >>
	reHeredocOverwr  = regexp.MustCompile(`>\s*/`)              // > /etc/...
	reBackgroundExec = regexp.MustCompile(`&\s*$`)              // в background — не контролируем
)

// hasDangerousPattern — глобальные опасные конструкции в строке команды.
func hasDangerousPattern(cmd string) bool {
	if rePipeShell.MatchString(cmd) {
		return true
	}
	// `>` и `>>` записывают в файл — опасно для системных файлов.
	if strings.Contains(cmd, " > ") || strings.Contains(cmd, " >> ") || reHeredocOverwr.MatchString(cmd) {
		_ = reRedirectWrite
		return true
	}
	if reBackgroundExec.MatchString(cmd) {
		return true
	}
	// Цепочки команд через ;, &&, || — может скрывать опасную команду в середине.
	// В безопасном whitelist таких нет.
	if strings.ContainsAny(cmd, ";&|") &&
		!strings.Contains(cmd, "&&") && // оставим простой случай foo && bar
		!strings.Contains(cmd, "||") {
		// одиночный | (пайп) — обычно безопасен (cmd | grep), но мы уже выловили pipe-to-shell.
		// Точку с запятой и фоновые задачи — опасно.
		if strings.Contains(cmd, ";") || reBackgroundExec.MatchString(cmd) {
			return true
		}
	}
	return false
}

// Classify определяет класс команды.
//
// Алгоритм:
//  1. Опасные паттерны (sudo, rm, > /etc/, cmd | sh, и т.д.) → Destructive.
//  2. Первый токен в blacklist → Destructive.
//  3. Первый токен в whitelist + destructive subcommand → Destructive.
//  4. Первый токен в whitelist → Safe.
//  5. Иначе → Unknown.
func Classify(command string) Class {
	cmd := strings.TrimSpace(command)
	if cmd == "" {
		return ClassUnknown
	}

	if hasDangerousPattern(cmd) {
		return ClassDestructive
	}

	tokens := tokenize(cmd)
	if len(tokens) == 0 {
		return ClassUnknown
	}
	first := tokens[0]

	// Поддержка `env VAR=val cmd ...` — берём настоящую команду после env.
	if first == "env" && len(tokens) >= 2 {
		// Пропускаем VAR=val токены.
		i := 1
		for i < len(tokens) && strings.Contains(tokens[i], "=") {
			i++
		}
		if i < len(tokens) {
			tokens = tokens[i:]
			first = tokens[0]
		}
	}

	if destructiveFirstTokens[first] {
		return ClassDestructive
	}
	if safeFirstTokens[first] {
		if isDestructiveSubcommand(tokens) {
			return ClassDestructive
		}
		return ClassSafe
	}
	return ClassUnknown
}

// tokenize — простой разбор на токены по пробелам с учётом кавычек.
// Не идеален для shell-парсинга (escape-последовательности, $vars), но достаточен
// чтобы определить первый токен и subcommand.
func tokenize(s string) []string {
	var out []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	for _, r := range s {
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case (r == ' ' || r == '\t') && !inSingle && !inDouble:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
