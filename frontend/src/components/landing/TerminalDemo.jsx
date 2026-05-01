import { useEffect, useRef, useState } from 'react';
import { motion } from 'framer-motion';
import Spinner from '../Spinner';

// Сценарий: юзер пишет wtf <вопрос> словами → агент через tool-use API
// сам выполняет read-only команды → находит причину → показывает destructive
// команду для ручного выполнения → ★ финальный ответ.
//
// Имитирует реальный quiet-режим wtf v2: каждая строка [HH:MM:SS], спиннер
// in-place во время выполнения команды, итог одной строкой после.
//
// Типы шагов:
//   user-cmd  — пользовательская команда в шелле (зелёная)
//   spin      — спиннер с reason+command (заменяется при следующем spin/done)
//   done      — закрытие спиннера (✓ или ✗) — зашёл в history
//   warn      — ⚠ destructive команда для юзера
//   warn-sub  — под-строка warn (с отступом)
//   final-h   — ★ ответ:
//   final     — текст финального ответа
//   blank     — пустая строка
//   pause     — пауза N мс

const session = [
  { t: 'user-cmd', text: 'wtf nginx не стартует' },
  { t: 'pause', ms: 600 },

  {
    t: 'spin',
    time: '03:55:40',
    en: 'Checking service status · systemctl status nginx -l',
    ru: 'Проверяю текущий статус сервиса · systemctl status nginx -l',
  },
  { t: 'pause', ms: 1100 },
  {
    t: 'done',
    ok: true,
    time: '03:55:40',
    en: 'Checking service status · systemctl status nginx -l · 17ms · 934B',
    ru: 'Проверяю текущий статус сервиса · systemctl status nginx -l · 17ms · 934B',
  },

  {
    t: 'spin',
    time: '03:55:41',
    en: 'Checking nginx config syntax · nginx -t',
    ru: 'Проверяю синтаксис конфигурации nginx · nginx -t',
  },
  { t: 'pause', ms: 900 },
  {
    t: 'done',
    ok: false,
    time: '03:55:41',
    en: 'Checking nginx config syntax · nginx -t · 12ms · exit=1 · 547B',
    ru: 'Проверяю синтаксис конфигурации nginx · nginx -t · 12ms · exit=1 · 547B',
  },
  {
    t: 'done-sub',
    en: '│ nginx: [emerg] cannot load certificate "/etc/letsencrypt/live/example.com/fullchain.pem"',
    ru: '│ nginx: [emerg] cannot load certificate "/etc/letsencrypt/live/example.com/fullchain.pem"',
  },
  {
    t: 'done-sub',
    en: '│ nginx: configuration file /etc/nginx/nginx.conf test failed',
    ru: '│ nginx: configuration file /etc/nginx/nginx.conf test failed',
  },

  {
    t: 'spin',
    time: '03:55:43',
    en: 'Checking certificate permissions · ls -ld /etc/letsencrypt/live',
    ru: 'Проверяю права на сертификаты · ls -ld /etc/letsencrypt/live',
  },
  { t: 'pause', ms: 800 },
  {
    t: 'done',
    ok: true,
    time: '03:55:43',
    en: 'Checking certificate permissions · ls -ld /etc/letsencrypt/live · 5ms · 78B',
    ru: 'Проверяю права на сертификаты · ls -ld /etc/letsencrypt/live · 5ms · 78B',
  },

  {
    t: 'spin',
    time: '03:55:45',
    en: 'Checking under what user master nginx is running · ps aux | grep nginx',
    ru: 'Под каким юзером запущен master nginx · ps aux | grep nginx',
  },
  { t: 'pause', ms: 700 },
  {
    t: 'done',
    ok: true,
    time: '03:55:45',
    en: 'Checking under what user master nginx is running · ps aux | grep nginx · 36ms · 249B',
    ru: 'Под каким юзером запущен master nginx · ps aux | grep nginx · 36ms · 249B',
  },

  { t: 'pause', ms: 600 },
  {
    t: 'warn',
    time: '03:55:55',
    en: 'run yourself (requires sudo / changes the system):',
    ru: 'выполни сам (требует sudo / меняет систему):',
  },
  {
    t: 'warn-sub',
    en: 'Set proper permissions on the cert directory so root nginx can read it.',
    ru: 'Установи права на директорию с сертификатом — nginx (root) сможет читать.',
  },
  {
    t: 'warn-cmd',
    text: '$ sudo chmod 750 /etc/letsencrypt/live && sudo systemctl restart nginx',
  },
  { t: 'pause', ms: 1400 },

  {
    t: 'final-h',
    time: '03:55:56',
    en: 'answer:',
    ru: 'ответ:',
  },
  { t: 'blank' },
  {
    t: 'final',
    en: 'nginx cannot start because the master process (running as root) does not',
    ru: 'nginx не стартует — мастер-процесс (root) не может прочитать SSL-сертификат',
  },
  {
    t: 'final',
    en: 'have read access to /etc/letsencrypt/live/example.com/fullchain.pem.',
    ru: '/etc/letsencrypt/live/example.com/fullchain.pem из-за прав на родительскую',
  },
  {
    t: 'final',
    en: 'The directory /etc/letsencrypt/live has mode 700 instead of 750.',
    ru: 'директорию /etc/letsencrypt/live (700 вместо 750).',
  },
  { t: 'blank' },
  {
    t: 'final',
    en: 'Run the command above — it fixes the permissions and restarts nginx.',
    ru: 'Выполни команду выше — она исправит права и перезапустит nginx.',
  },
  { t: 'pause', ms: 2800 },

  {
    t: 'log',
    en: '— wtf is on call. describe the problem, agent does the rest.',
    ru: '— wtf на связи. опиши проблему — агент сам разберётся.',
  },
];

const TYPE_DELAY_FIRST = 500;
const TYPE_DELAY = 200;
const MAX_LINES = 22;

const lineClass = (t) => {
  switch (t) {
    case 'user-cmd':
      return 'text-zinc-100';
    case 'spin':
      return 'text-amber-300';
    case 'done':
      return 'text-amber-400';
    case 'done-sub':
      return 'text-zinc-500';
    case 'warn':
      return 'text-amber-300';
    case 'warn-sub':
      return 'text-zinc-400';
    case 'warn-cmd':
      return 'text-amber-300/90 font-medium';
    case 'final-h':
      return 'text-amber-400 font-semibold';
    case 'final':
      return 'text-zinc-200';
    case 'log':
      return 'text-zinc-500 italic';
    default:
      return 'text-zinc-500';
  }
};

const resolveText = (step, lang) => {
  if (step.t === 'blank') return ' ';
  return step.text ?? step[lang] ?? step.en;
};

// renderLine — стилизация одной строки. Префикс [HH:MM:SS] + иконка + текст,
// 14-символьный отступ для under-строк (как в реальном wtf).
function renderLine({ step, lang }) {
  const text = resolveText(step, lang);

  if (step.t === 'user-cmd') {
    return (
      <>
        <span className="text-amber-400">admingod@srv:~$ </span>
        <span>{text}</span>
      </>
    );
  }

  if (step.t === 'done') {
    const icon = step.ok ? '✓' : '✗';
    const iconColor = step.ok ? 'text-amber-400' : 'text-red-400';
    return (
      <>
        <span className="text-zinc-500">[{step.time}]</span>{' '}
        <span className={iconColor}>{icon}</span> <span>{text}</span>
      </>
    );
  }

  if (step.t === 'done-sub') {
    return <span className="ml-[6.5rem] text-zinc-500">{text}</span>;
  }

  if (step.t === 'warn') {
    return (
      <>
        <span className="text-zinc-500">[{step.time}]</span>{' '}
        <span className="text-amber-300">⚠</span>{' '}
        <span className="text-amber-300">{text}</span>
      </>
    );
  }

  if (step.t === 'warn-sub') {
    return <span className="ml-[6.5rem] text-zinc-400">{text}</span>;
  }

  if (step.t === 'warn-cmd') {
    return <span className="ml-[6.5rem] text-amber-300/90">{text}</span>;
  }

  if (step.t === 'final-h') {
    return (
      <>
        <span className="text-zinc-500">[{step.time}]</span>{' '}
        <span className="text-amber-400">★</span>{' '}
        <span className="text-amber-400 font-semibold">{text}</span>
      </>
    );
  }

  if (step.t === 'final') {
    return <span className="ml-[6.5rem]">{text}</span>;
  }

  if (step.t === 'log') {
    return <span>{text}</span>;
  }

  return <span>{text}</span>;
}

export default function TerminalDemo() {
  const [lang, setLang] = useState('ru');
  const [history, setHistory] = useState([]);
  const [activeSpin, setActiveSpin] = useState(null);
  const [idx, setIdx] = useState(0);
  const keyRef = useRef(0);
  const scrollRef = useRef(null);

  useEffect(() => {
    if (idx >= session.length) return;
    const step = session[idx];

    if (step.t === 'pause') {
      const id = setTimeout(() => setIdx((i) => i + 1), step.ms);
      return () => clearTimeout(id);
    }

    const delay = idx === 0 ? TYPE_DELAY_FIRST : TYPE_DELAY;
    const id = setTimeout(() => {
      if (step.t === 'spin') {
        setActiveSpin({ step, key: keyRef.current++ });
      } else if (step.t === 'done') {
        setActiveSpin(null);
        setHistory((h) => trim([...h, { step, key: keyRef.current++ }]));
      } else {
        setActiveSpin(null);
        setHistory((h) => trim([...h, { step, key: keyRef.current++ }]));
      }
      setIdx((i) => i + 1);
    }, delay);
    return () => clearTimeout(id);
  }, [idx]);

  useEffect(() => {
    if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
  }, [history, activeSpin]);

  return (
    <div className="rounded-2xl bg-zinc-900/80 border border-zinc-800 overflow-hidden shadow-2xl shadow-amber-400/5 backdrop-blur">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-zinc-800 bg-zinc-900/60">
        <div className="w-3 h-3 rounded-full bg-zinc-700" />
        <div className="w-3 h-3 rounded-full bg-zinc-700" />
        <div className="w-3 h-3 rounded-full bg-zinc-700" />
        <span className="ml-3 text-xs text-zinc-500 font-mono">admingod@srv:~</span>
        <div className="ml-auto flex items-center gap-1 text-[10px] font-mono">
          <button
            type="button"
            onClick={() => setLang('ru')}
            className={`px-2 py-0.5 rounded transition-colors ${
              lang === 'ru' ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-500 hover:text-zinc-300'
            }`}
            aria-pressed={lang === 'ru'}
          >
            RU
          </button>
          <button
            type="button"
            onClick={() => setLang('en')}
            className={`px-2 py-0.5 rounded transition-colors ${
              lang === 'en' ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-500 hover:text-zinc-300'
            }`}
            aria-pressed={lang === 'en'}
          >
            EN
          </button>
        </div>
      </div>

      <div ref={scrollRef} className="p-5 font-mono text-xs md:text-sm leading-relaxed h-[420px] overflow-hidden">
        {history.map(({ step, key }) => (
          <motion.div
            key={key}
            initial={{ opacity: 0, x: -8 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.4, ease: 'easeOut' }}
            className={lineClass(step.t)}
          >
            {renderLine({ step, lang })}
          </motion.div>
        ))}

        {activeSpin && (
          <motion.div
            key={activeSpin.key}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.2 }}
            className={lineClass('spin')}
          >
            <span className="text-zinc-500">[{activeSpin.step.time}]</span>{' '}
            <Spinner className="mr-1.5 text-amber-400 inline-block" />
            <span>{resolveText(activeSpin.step, lang)}</span>
          </motion.div>
        )}

        {!activeSpin && (
          <span className="inline-block w-2 h-4 bg-amber-400 animate-pulse ml-0.5 align-middle" />
        )}
      </div>
    </div>
  );
}

function trim(arr) {
  return arr.length > MAX_LINES ? arr.slice(arr.length - MAX_LINES) : arr;
}
