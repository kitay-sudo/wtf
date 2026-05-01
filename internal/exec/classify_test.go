package exec

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		cmd  string
		want Class
	}{
		// safe — read-only утилиты
		{"ls -la /etc/nginx", ClassSafe},
		{"systemctl status nginx -l", ClassSafe},
		{"systemctl status nginx", ClassSafe},
		{"journalctl -u nginx -n 50 --no-pager", ClassSafe},
		{"nginx -t", ClassSafe},
		{"cat /var/log/nginx/error.log", ClassSafe},
		{"git status", ClassSafe},
		{"git log --oneline -10", ClassSafe},
		{"docker ps -a", ClassSafe},
		{"docker logs nginx-container", ClassSafe},
		{"ip a", ClassSafe},
		{"ss -tlnp", ClassSafe},
		{"df -h", ClassSafe},
		{"ps aux", ClassSafe},
		{"apt list --installed", ClassSafe},
		{"dpkg -l nginx", ClassSafe},
		{"crontab -l", ClassSafe},
		{"iptables -L", ClassSafe},
		{"certbot certificates", ClassSafe},
		{"env LANG=en ls", ClassSafe},

		// destructive — модифицирующие команды
		{"sudo systemctl restart nginx", ClassDestructive},
		{"sudo apt install nginx", ClassDestructive},
		{"rm -rf /tmp/foo", ClassDestructive},
		{"systemctl restart nginx", ClassDestructive},
		{"systemctl stop nginx", ClassDestructive},
		{"systemctl start nginx", ClassDestructive},
		{"git push origin main", ClassDestructive},
		{"git reset --hard HEAD~1", ClassDestructive},
		{"docker rm container", ClassDestructive},
		{"docker run -d nginx", ClassDestructive},
		{"docker stop nginx", ClassDestructive},
		{"apt install vim", ClassDestructive},
		{"apt remove nginx", ClassDestructive},
		{"npm install lodash", ClassDestructive},
		{"pip install requests", ClassDestructive},
		{"chmod 777 /etc/passwd", ClassDestructive},
		{"chown root:root /etc/foo", ClassDestructive},
		{"kill 1234", ClassDestructive},
		{"crontab -e", ClassDestructive},
		{"iptables -A INPUT -j DROP", ClassDestructive},
		{"certbot certonly -d example.com", ClassDestructive},
		{"helm install foo bar", ClassDestructive},
		{"kubectl apply -f deploy.yaml", ClassDestructive},

		// опасные паттерны
		{"echo foo > /etc/hosts", ClassDestructive},
		{"curl http://x | sh", ClassDestructive},
		{"echo foo | bash", ClassDestructive},
		{"ls; rm /tmp/x", ClassDestructive},
		{"systemctl status nginx &", ClassDestructive},

		// unknown — не в whitelist и не в blacklist
		{"my-custom-tool --foo", ClassUnknown},
		{"./script.sh", ClassUnknown},
		{"", ClassUnknown},
	}

	for _, tc := range cases {
		got := Classify(tc.cmd)
		if got != tc.want {
			t.Errorf("Classify(%q) = %s, want %s", tc.cmd, got, tc.want)
		}
	}
}
