import { useEffect, useRef, useState } from 'react';
import { motion } from 'framer-motion';

// Несколько реалистичных сценариев. Юзер видит, как падает команда,
// потом запускает wtf, и сразу — короткое объяснение и фикс.
const session = [
  { t: 'cmd', text: '$ npm run build' },
  {
    t: 'err',
    en: '> wtf-frontend@0.1.0 build',
    ru: '> wtf-frontend@0.1.0 build',
  },
  { t: 'err', text: '> vite build' },
  { t: 'err', text: '' },
  {
    t: 'err',
    text: 'error during build:',
  },
  {
    t: 'err',
    text: 'Error: Cannot find module \'./pages/Landing\' imported from src/App.jsx',
  },
  { t: 'err', text: '    at finalizeResolution (node:internal/modules/esm/resolve:265)' },
  { t: 'err', text: '    at moduleResolve (node:internal/modules/esm/resolve:933)' },
  { t: 'pause', ms: 1200 },

  { t: 'cmd', text: '$ wtf' },
  {
    t: 'spin',
    en: 'analyzing via claude (haiku-4-5)...',
    ru: 'анализ через claude (haiku-4-5)...',
  },
  { t: 'pause', ms: 700 },
  {
    t: 'ok',
    en: '✓ answer from claude',
    ru: '✓ ответ от claude',
  },
  { t: 'blank' },
  {
    t: 'ans-h',
    en: 'What happened:',
    ru: 'Что случилось:',
  },
  {
    t: 'ans',
    en: 'Vite не находит файл src/pages/Landing.jsx — путь импорта не совпадает с реальным.',
    ru: 'Vite не находит файл src/pages/Landing.jsx — путь импорта не совпадает с реальным.',
  },
  { t: 'blank' },
  { t: 'ans-h', en: 'Fix:', ru: 'Как починить:' },
  { t: 'ans', en: '1) Проверь, что файл существует:', ru: '1) Проверь, что файл существует:' },
  { t: 'code', text: '   ls src/pages/Landing.jsx' },
  { t: 'ans', en: '2) Если расширение .jsx — Vite требует его явно:', ru: '2) Если расширение .jsx — Vite требует его явно:' },
  { t: 'code', text: "   import Landing from './pages/Landing.jsx'" },
  { t: 'ans', en: '3) Или добавь resolve.extensions в vite.config.js', ru: '3) Или добавь resolve.extensions в vite.config.js' },
  { t: 'pause', ms: 1500 },

  { t: 'cmd', text: '$ docker compose up' },
  { t: 'err', text: 'Error response from daemon: bind: address already in use' },
  { t: 'err', text: 'Error: failed to start container "web": port 5432 is allocated' },
  { t: 'pause', ms: 1000 },

  { t: 'cmd', text: '$ wtf' },
  {
    t: 'spin',
    en: 'analyzing via claude (haiku-4-5)...',
    ru: 'анализ через claude (haiku-4-5)...',
  },
  { t: 'pause', ms: 600 },
  { t: 'ok', en: '✓ answer from claude', ru: '✓ ответ от claude' },
  { t: 'blank' },
  {
    t: 'ans-h',
    en: 'What happened:',
    ru: 'Что случилось:',
  },
  {
    t: 'ans',
    en: 'Порт 5432 уже занят другим процессом — старая Postgres всё ещё запущена.',
    ru: 'Порт 5432 уже занят другим процессом — старая Postgres всё ещё запущена.',
  },
  { t: 'blank' },
  { t: 'ans-h', en: 'Fix:', ru: 'Как починить:' },
  { t: 'ans', en: '1) Найти кто держит порт:', ru: '1) Найти кто держит порт:' },
  { t: 'code', text: '   sudo lsof -iTCP:5432 -sTCP:LISTEN' },
  { t: 'ans', en: '2) Остановить системный postgres:', ru: '2) Остановить системный postgres:' },
  { t: 'code', text: '   sudo systemctl stop postgresql' },
  { t: 'ans', en: '3) Или поменять порт в docker-compose.yml: 5433:5432', ru: '3) Или поменять порт в docker-compose.yml: 5433:5432' },
  { t: 'pause', ms: 2000 },

  { t: 'cmd', text: '$ git push origin main' },
  { t: 'err', text: 'To github.com:user/repo.git' },
  { t: 'err', text: '! [rejected]        main -> main (non-fast-forward)' },
  { t: 'err', text: 'error: failed to push some refs' },
  { t: 'err', text: "hint: Updates were rejected because the tip of your current branch is behind" },
  { t: 'pause', ms: 1000 },

  { t: 'cmd', text: '$ wtf' },
  { t: 'spin', en: 'analyzing via claude...', ru: 'анализ через claude...' },
  { t: 'pause', ms: 500 },
  { t: 'ok', en: '✓ answer from claude', ru: '✓ ответ от claude' },
  { t: 'blank' },
  { t: 'ans-h', en: 'What happened:', ru: 'Что случилось:' },
  {
    t: 'ans',
    en: 'На удалённой ветке появились коммиты, которых нет локально — нужно сначала их подтянуть.',
    ru: 'На удалённой ветке появились коммиты, которых нет локально — нужно сначала их подтянуть.',
  },
  { t: 'blank' },
  { t: 'ans-h', en: 'Fix:', ru: 'Как починить:' },
  { t: 'ans', en: '1) Безопасно — rebase на удалённый main:', ru: '1) Безопасно — rebase на удалённый main:' },
  { t: 'code', text: '   git pull --rebase origin main && git push' },
  { t: 'ans', en: '2) Если знаешь что делаешь — force-with-lease:', ru: '2) Если знаешь что делаешь — force-with-lease:' },
  { t: 'code', text: '   git push --force-with-lease' },
  { t: 'pause', ms: 2200 },

  {
    t: 'log',
    en: '— wtf is watching. write any failed command, then `wtf`.',
    ru: '— wtf наготове. упади и напиши `wtf`.',
  },
];

const TYPE_DELAY_FIRST = 500;
const TYPE_DELAY = 280;
const MAX_LINES = 18;

const lineClass = (t) =>
  t === 'cmd'
    ? 'text-zinc-100'
    : t === 'ok'
    ? 'text-amber-400'
    : t === 'err'
    ? 'text-red-400/90'
    : t === 'spin'
    ? 'text-amber-300'
    : t === 'ans-h'
    ? 'text-amber-400 font-semibold'
    : t === 'ans'
    ? 'text-zinc-200'
    : t === 'code'
    ? 'text-amber-300/90 font-medium'
    : t === 'log'
    ? 'text-zinc-500 italic'
    : 'text-zinc-500';

const resolveText = (step, lang) => {
  if (step.t === 'blank') return ' ';
  return step.text ?? step[lang] ?? step.en;
};

export default function TerminalDemo() {
  const [lang, setLang] = useState('ru');
  const [history, setHistory] = useState([]);
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
      setHistory((h) => {
        const next = [...h, { step, key: keyRef.current++ }];
        return next.length > MAX_LINES ? next.slice(next.length - MAX_LINES) : next;
      });
      setIdx((i) => i + 1);
    }, delay);
    return () => clearTimeout(id);
  }, [idx]);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [history]);

  return (
    <div className="rounded-2xl bg-zinc-900/80 border border-zinc-800 overflow-hidden shadow-2xl shadow-amber-400/5 backdrop-blur">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-zinc-800 bg-zinc-900/60">
        <div className="w-3 h-3 rounded-full bg-zinc-700" />
        <div className="w-3 h-3 rounded-full bg-zinc-700" />
        <div className="w-3 h-3 rounded-full bg-zinc-700" />
        <span className="ml-3 text-xs text-zinc-500 font-mono">~/projects/myapp</span>
        <div className="ml-auto flex items-center gap-1 text-[10px] font-mono">
          <button
            type="button"
            onClick={() => setLang('ru')}
            className={`px-2 py-0.5 rounded transition-colors ${
              lang === 'ru'
                ? 'bg-zinc-800 text-zinc-100'
                : 'text-zinc-500 hover:text-zinc-300'
            }`}
            aria-pressed={lang === 'ru'}
          >
            RU
          </button>
          <button
            type="button"
            onClick={() => setLang('en')}
            className={`px-2 py-0.5 rounded transition-colors ${
              lang === 'en'
                ? 'bg-zinc-800 text-zinc-100'
                : 'text-zinc-500 hover:text-zinc-300'
            }`}
            aria-pressed={lang === 'en'}
          >
            EN
          </button>
        </div>
      </div>

      <div
        ref={scrollRef}
        className="p-5 font-mono text-xs md:text-sm leading-relaxed h-[360px] overflow-hidden"
      >
        {history.map(({ step, key }) => (
          <motion.div
            key={key}
            initial={{ opacity: 0, x: -8 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.45, ease: 'easeOut' }}
            className={lineClass(step.t)}
          >
            {step.t === 'spin' && (
              <span className="inline-block mr-1.5 animate-spin-slow">⠋</span>
            )}
            {resolveText(step, lang)}
          </motion.div>
        ))}
        <span className="inline-block w-2 h-4 bg-amber-400 animate-pulse ml-0.5 align-middle" />
      </div>
    </div>
  );
}
