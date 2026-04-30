import { useEffect, useRef, useState } from 'react';
import { motion } from 'framer-motion';
import Spinner from '../Spinner';

// Сценарий: команда падает → юзер пишет wtf → бегущая последовательность статусов
// (читает stderr, собирает контекст, чистит секреты, шлёт в API, ждёт, парсит) →
// готовый Markdown-ответ. Похоже на то, что делает реальная утилита.
//
// Типы шагов:
//   cmd     — пользовательская команда
//   err     — строка из stderr (красная)
//   spin    — статус-строка со спиннером, ЗАМЕНЯЕТСЯ при следующем 'spin' (in-place)
//   ok      — финальный успех (зелёная-жёлтая галка), снимает спиннер
//   blank   — пустая строка
//   ans-h   — заголовок секции ответа
//   ans     — текст ответа
//   code    — код-строка (отступ + цвет)
//   log     — нейтральный комментарий (серый, курсив)
//   pause   — пауза N мс (не выводится)

const session = [
  { t: 'cmd', text: '$ npm run build' },
  { t: 'err', text: '> wtf-frontend@0.1.0 build' },
  { t: 'err', text: '> vite build' },
  { t: 'err', text: '' },
  { t: 'err', text: 'error during build:' },
  { t: 'err', text: "Error: Cannot find module './pages/Landing' imported from src/App.jsx" },
  { t: 'err', text: '    at finalizeResolution (node:internal/modules/esm/resolve:265)' },
  { t: 'pause', ms: 1000 },

  { t: 'cmd', text: '$ wtf' },
  { t: 'spin', en: 'Reading stderr...',                  ru: 'Читаю stderr...' },
  { t: 'pause', ms: 350 },
  { t: 'spin', en: 'Collecting context (OS, shell, git)...', ru: 'Собираю контекст (OS, shell, git)...' },
  { t: 'pause', ms: 400 },
  { t: 'spin', en: 'Redacting secrets...',                ru: 'Чищу секреты (regex × 13)...' },
  { t: 'pause', ms: 350 },
  { t: 'spin', en: 'Calling claude (haiku-4-5)...',       ru: 'Запрос в claude (haiku-4-5)...' },
  { t: 'pause', ms: 700 },
  { t: 'spin', en: 'Receiving stream...',                 ru: 'Получаю ответ...' },
  { t: 'pause', ms: 500 },
  { t: 'ok',   en: 'Done · 1.2s · cached for 30 days',    ru: 'Готово · 1.2с · сохранено в кеш на 30 дней' },
  { t: 'blank' },

  { t: 'ans-h', en: 'What happened:', ru: 'Что случилось:' },
  {
    t: 'ans',
    en: 'Vite cannot find src/pages/Landing.jsx — import path does not match the file.',
    ru: 'Vite не находит src/pages/Landing.jsx — путь импорта не совпадает с реальным.',
  },
  { t: 'blank' },
  { t: 'ans-h', en: 'Fix:', ru: 'Как починить:' },
  { t: 'ans', en: '1) Check the file exists:', ru: '1) Проверь, что файл существует:' },
  { t: 'code', text: '   ls src/pages/Landing.jsx' },
  { t: 'ans', en: '2) Vite needs the .jsx extension explicitly:', ru: '2) Vite требует расширение .jsx явно:' },
  { t: 'code', text: "   import Landing from './pages/Landing.jsx'" },
  { t: 'ans', en: '3) Or set resolve.extensions in vite.config.js', ru: '3) Или добавь resolve.extensions в vite.config.js' },
  { t: 'pause', ms: 1700 },

  { t: 'cmd', text: '$ docker compose up' },
  { t: 'err', text: 'Error response from daemon: bind: address already in use' },
  { t: 'err', text: 'Error: failed to start container "web": port 5432 is allocated' },
  { t: 'pause', ms: 800 },

  { t: 'cmd', text: '$ wtf' },
  { t: 'spin', en: 'Reading stderr...',           ru: 'Читаю stderr...' },
  { t: 'pause', ms: 300 },
  { t: 'spin', en: 'Cache hit — returning instantly', ru: 'Найдено в кеше — отвечаю мгновенно' },
  { t: 'pause', ms: 400 },
  { t: 'ok',   en: 'Done · 0.04s · from cache',   ru: 'Готово · 0.04с · из кеша' },
  { t: 'blank' },

  { t: 'ans-h', en: 'What happened:', ru: 'Что случилось:' },
  {
    t: 'ans',
    en: 'Port 5432 is busy — system Postgres is still running.',
    ru: 'Порт 5432 занят — старая Postgres всё ещё запущена.',
  },
  { t: 'blank' },
  { t: 'ans-h', en: 'Fix:', ru: 'Как починить:' },
  { t: 'ans', en: '1) Find who holds the port:', ru: '1) Кто держит порт:' },
  { t: 'code', text: '   sudo lsof -iTCP:5432 -sTCP:LISTEN' },
  { t: 'ans', en: '2) Stop system postgres:', ru: '2) Остановить системный postgres:' },
  { t: 'code', text: '   sudo systemctl stop postgresql' },
  { t: 'ans', en: '3) Or change the port in docker-compose.yml: 5433:5432', ru: '3) Или поменять порт: 5433:5432' },
  { t: 'pause', ms: 1800 },

  { t: 'cmd', text: '$ git push origin main' },
  { t: 'err', text: '! [rejected]        main -> main (non-fast-forward)' },
  { t: 'err', text: 'error: failed to push some refs' },
  { t: 'pause', ms: 700 },

  { t: 'cmd', text: '$ wtf' },
  { t: 'spin', en: 'Reading stderr...',                  ru: 'Читаю stderr...' },
  { t: 'pause', ms: 280 },
  { t: 'spin', en: 'Calling claude...',                  ru: 'Запрос в claude...' },
  { t: 'pause', ms: 600 },
  { t: 'ok',   en: 'Done · 0.9s',                        ru: 'Готово · 0.9с' },
  { t: 'blank' },

  { t: 'ans-h', en: 'What happened:', ru: 'Что случилось:' },
  {
    t: 'ans',
    en: 'Remote has commits you do not have locally — fast-forward not possible.',
    ru: 'На удалённой ветке есть коммиты, которых нет локально — fast-forward невозможен.',
  },
  { t: 'blank' },
  { t: 'ans-h', en: 'Fix:', ru: 'Как починить:' },
  { t: 'ans', en: '1) Safe — rebase on top of remote:', ru: '1) Безопасно — rebase на удалённый main:' },
  { t: 'code', text: '   git pull --rebase origin main && git push' },
  { t: 'ans', en: '2) If you know what you are doing — force-with-lease:', ru: '2) Если знаешь что делаешь — force-with-lease:' },
  { t: 'code', text: '   git push --force-with-lease' },
  { t: 'pause', ms: 2200 },

  { t: 'log', en: '— wtf is watching. fail any command, then `wtf`.', ru: '— wtf наготове. упади любой командой, потом `wtf`.' },
];

const TYPE_DELAY_FIRST = 500;
const TYPE_DELAY = 220;
const MAX_LINES = 18;

const lineClass = (t) =>
  t === 'cmd' ? 'text-zinc-100'
    : t === 'ok' ? 'text-amber-400'
    : t === 'err' ? 'text-red-400/90'
    : t === 'spin' ? 'text-amber-300'
    : t === 'ans-h' ? 'text-amber-400 font-semibold'
    : t === 'ans' ? 'text-zinc-200'
    : t === 'code' ? 'text-amber-300/90 font-medium'
    : t === 'log' ? 'text-zinc-500 italic'
    : 'text-zinc-500';

const resolveText = (step, lang) => {
  if (step.t === 'blank') return ' ';
  return step.text ?? step[lang] ?? step.en;
};

export default function TerminalDemo() {
  const [lang, setLang] = useState('ru');
  const [history, setHistory] = useState([]);
  const [activeSpin, setActiveSpin] = useState(null); // {step, key} | null
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
        // Спиннер — это ОДНА живая строка, которая обновляется in-place.
        // Не уходит в history до тех пор, пока не закроется ok/err.
        setActiveSpin({ step, key: keyRef.current++ });
      } else if (step.t === 'ok' || step.t === 'err') {
        // Закрытие спиннера — снимаем активный, добавляем строку в history.
        setActiveSpin(null);
        setHistory((h) => trim([...h, { step, key: keyRef.current++ }]));
      } else {
        // Обычная строка — без активного спиннера.
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
        <span className="ml-3 text-xs text-zinc-500 font-mono">~/projects/myapp</span>
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

      <div ref={scrollRef} className="p-5 font-mono text-xs md:text-sm leading-relaxed h-[360px] overflow-hidden">
        {history.map(({ step, key }) => (
          <motion.div
            key={key}
            initial={{ opacity: 0, x: -8 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.4, ease: 'easeOut' }}
            className={lineClass(step.t)}
          >
            {resolveText(step, lang)}
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
            <Spinner className="mr-1.5 text-amber-400" />
            {resolveText(activeSpin.step, lang)}
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
