import { motion, AnimatePresence } from 'framer-motion';
import {
  Terminal,
  Zap,
  Brain,
  ShieldCheck,
  Lock,
  Github,
  Cpu,
  Globe,
  Copy,
  Check,
  ArrowUp,
  History,
  Sparkles,
  Layers,
  KeyRound,
  Heart,
  Send,
  Server,
  ArrowRight,
} from 'lucide-react';
import { useState, useEffect } from 'react';
import GridBackground from '../components/landing/GridBackground';
import Reveal from '../components/landing/Reveal';
import TerminalDemo from '../components/landing/TerminalDemo';
import FeatureCard from '../components/landing/FeatureCard';
import FAQItem from '../components/landing/FAQItem';
import WtfMark from '../components/landing/WtfMark';
import AnnouncementBar from '../components/landing/AnnouncementBar';
import SectionDivider from '../components/landing/SectionDivider';

const REPO_URL = 'https://github.com/kitay-sudo/wtf';
const INSTALL_CMD_UNIX = 'curl -sSL https://raw.githubusercontent.com/kitay-sudo/wtf/main/install.sh | sudo bash';
const INSTALL_CMD_BREW = 'brew install kitay-sudo/wtf/wtf';
const INSTALL_CMD_PWSH = 'iwr -useb https://raw.githubusercontent.com/kitay-sudo/wtf/main/install.ps1 | iex';

export default function Landing() {
  return (
    <div className="min-h-dvh bg-zinc-950 text-zinc-100 antialiased">
      <AnnouncementBar />
      <Nav />
      <Hero />
      <WhyName />
      <LogosStrip />
      <Features />
      <HowItWorks />
      <DemoSection />
      <FAQ />
      <Support />
      <CTA />
      <Changelog />
      <Footer />
      <BackToTop />
    </div>
  );
}

function BackToTop() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const onScroll = () => setVisible(window.scrollY > 400);
    onScroll();
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  const scrollUp = () => window.scrollTo({ top: 0, behavior: 'smooth' });

  return (
    <AnimatePresence>
      {visible && (
        <motion.button
          key="back-to-top"
          type="button"
          onClick={scrollUp}
          aria-label="Наверх"
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 12 }}
          transition={{ duration: 0.2, ease: 'easeOut' }}
          className="fixed z-50 bottom-6 right-6 inline-flex items-center justify-center w-11 h-11 rounded-full border border-amber-400/30 bg-zinc-900/80 backdrop-blur text-amber-400 hover:text-amber-300 hover:border-amber-400/60 hover:bg-zinc-900 shadow-lg shadow-amber-400/10 transition-colors"
        >
          <ArrowUp size={18} strokeWidth={2.2} />
        </motion.button>
      )}
    </AnimatePresence>
  );
}

function Nav() {
  return (
    <header className="sticky top-0 z-50 border-b border-zinc-900/80 bg-zinc-950/70 backdrop-blur-lg">
      <div className="max-w-6xl mx-auto px-5 h-14 flex items-center justify-between">
        <a href="#top" className="flex items-center gap-2.5 font-semibold">
          <WtfMark size={28} />
          <span className="tracking-tight">wtf</span>
        </a>

        <nav className="hidden md:flex items-center gap-7 text-sm text-zinc-400">
          <a href="#why" className="hover:text-zinc-100 transition-colors">Зачем</a>
          <a href="#features" className="hover:text-zinc-100 transition-colors">Возможности</a>
          <a href="#how" className="hover:text-zinc-100 transition-colors">Как работает</a>
          <a href="#install" className="hover:text-zinc-100 transition-colors">Установка</a>
          <a href="#faq" className="hover:text-zinc-100 transition-colors">FAQ</a>
          <a href="#changelog" className="hover:text-zinc-100 transition-colors">Изменения</a>
          <a href="#support" className="text-amber-300/90 hover:text-amber-200 transition-colors inline-flex items-center gap-1.5">
            <Heart size={12} fill="currentColor" />
            Стена чести
          </a>
        </nav>

        <a
          href={REPO_URL}
          target="_blank"
          rel="noreferrer"
          className="text-sm font-medium bg-amber-400 hover:bg-amber-300 text-zinc-950 rounded-lg px-3.5 py-1.5 transition-colors flex items-center gap-1.5"
        >
          <Github size={14} />
          GitHub
        </a>
      </div>
    </header>
  );
}

function Hero() {
  return (
    <section id="top" className="relative overflow-hidden">
      <GridBackground />

      <div className="relative max-w-6xl mx-auto px-5 pt-20 md:pt-28 pb-20 md:pb-32">
        <Reveal>
          <div className="flex justify-center">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-zinc-800 bg-zinc-900/60 text-xs text-zinc-400">
              <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
              <span className="font-mono text-zinc-500">v0.1</span>
              <span className="h-3 w-px bg-zinc-700" />
              CLI · Claude · OpenAI · Gemini · MIT
            </div>
          </div>
        </Reveal>

        <Reveal delay={0.05}>
          <h1 className="mt-6 text-center text-4xl md:text-6xl font-semibold tracking-tight leading-[1.05]">
            Не понял что в терминале — пиши <br />
            <span className="inline-flex items-center gap-3 align-middle">
              <span className="bg-gradient-to-r from-amber-400 to-yellow-300 bg-clip-text text-transparent">
                wtf
              </span>
              <WtfMark size={56} />
            </span>
          </h1>
        </Reveal>

        <Reveal delay={0.08}>
          <p className="mt-6 text-center text-lg md:text-xl text-zinc-200 max-w-2xl mx-auto leading-relaxed font-medium">
            Любой непонятный вывод — упавший stack trace, status сервиса, дамп конфига, странный JSON, лог nginx — пиши <code className="font-mono text-amber-300">wtf</code> и получай человеческое объяснение прямо в терминале. Если есть что чинить — ещё и 2-3 готовых действия.
          </p>
        </Reveal>

        <Reveal delay={0.16}>
          <p className="mt-5 text-center text-sm md:text-base text-zinc-400 max-w-2xl mx-auto leading-relaxed">
            Один Go-бинарник. Три AI-провайдера на выбор: Claude, OpenAI, Gemini —
            твой ключ, твой счёт. Локальный кеш, полная редакция секретов перед отправкой.
            Open-source, MIT.
          </p>
        </Reveal>

        <Reveal delay={0.15}>
          <div className="mt-8 max-w-2xl mx-auto">
            <InstallCommand />
          </div>
        </Reveal>

        <Reveal delay={0.2}>
          <div className="mt-5 flex flex-col sm:flex-row items-center justify-center gap-3">
            <a
              href={REPO_URL}
              target="_blank"
              rel="noreferrer"
              className="w-full sm:w-auto flex items-center justify-center gap-2 text-zinc-300 hover:text-zinc-100 border border-zinc-800 hover:border-zinc-700 rounded-xl px-5 py-3 transition-colors"
            >
              <Github size={16} />
              Исходники на GitHub
            </a>
            <a
              href="#how"
              className="w-full sm:w-auto flex items-center justify-center gap-2 text-zinc-400 hover:text-zinc-200 transition-colors px-5 py-3"
            >
              <Terminal size={16} />
              Как работает
            </a>
          </div>
        </Reveal>

        <Reveal delay={0.25}>
          <div className="mt-6 text-center text-xs text-zinc-500">
            Бесплатно · MIT-лицензия · macOS / Linux / Windows · работает с любым языком и тулом
          </div>
        </Reveal>

        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.3, ease: [0.22, 1, 0.36, 1] }}
          className="mt-14 md:mt-20 max-w-3xl mx-auto"
        >
          <TerminalDemo />
        </motion.div>
      </div>
    </section>
  );
}

function InstallCommand() {
  const tabs = [
    { id: 'unix', label: 'curl', cmd: INSTALL_CMD_UNIX },
    { id: 'brew', label: 'brew', cmd: INSTALL_CMD_BREW },
    { id: 'pwsh', label: 'pwsh', cmd: INSTALL_CMD_PWSH },
  ];
  const [tab, setTab] = useState('unix');
  const [copied, setCopied] = useState(false);
  const active = tabs.find((t) => t.id === tab);

  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(active.cmd);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      /* empty */
    }
  };

  return (
    <div className="rounded-2xl border border-amber-400/30 bg-zinc-900/80 backdrop-blur shadow-lg shadow-amber-400/10 overflow-hidden">
      <div className="flex items-center border-b border-zinc-800/80 px-2 pt-2 gap-1">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`px-3 py-1.5 text-xs font-mono rounded-t-md transition-colors ${
              tab === t.id
                ? 'bg-zinc-800/80 text-amber-300 border-b-0'
                : 'text-zinc-500 hover:text-zinc-300'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>
      <div className="flex items-center justify-between gap-3 p-4">
        <code className="text-xs sm:text-sm text-amber-300 font-mono break-all flex-1 min-w-0">
          {active.cmd}
        </code>
        <button
          onClick={onCopy}
          className="shrink-0 inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1.5 rounded-md border border-zinc-700 hover:border-amber-400/50 hover:bg-amber-400/10 transition-colors text-zinc-300"
          aria-label="Скопировать команду"
        >
          {copied ? <Check size={14} className="text-amber-400" /> : <Copy size={14} />}
          {copied ? 'Скопировано' : 'Копировать'}
        </button>
      </div>
    </div>
  );
}

function WhyName() {
  return (
    <section id="why" className="relative border-y border-zinc-900/80 overflow-hidden">
      <div className="relative max-w-4xl mx-auto px-5 py-20 md:py-28 text-center">
        <Reveal>
          <SectionDivider symbol="?" label="The Why" />
          <h2 className="text-3xl md:text-5xl font-semibold tracking-tight leading-tight">
            Почему <span className="text-amber-400">wtf</span>?
          </h2>
        </Reveal>

        <Reveal delay={0.1}>
          <p className="mt-6 text-zinc-400 leading-relaxed md:text-lg">
            Это первое, что ты пишешь, когда в терминале что-то непонятное —
            упавшая команда, длинный статус сервиса, дамп конфига, незнакомый JSON.
            Раньше — копировал в Google или ChatGPT, тратил минуты.
            Теперь — просто <code className="font-mono text-amber-300">wtf</code>.
          </p>
        </Reveal>

        <Reveal delay={0.15}>
          <p className="mt-4 text-zinc-400 leading-relaxed md:text-lg">
            Утилита читает последний вывод терминала (stdout + stderr), добавляет контекст
            (ОС, shell, package manager, git branch), вычищает секреты и спрашивает у AI.
            На выходе — короткое объяснение того, что произошло, и при необходимости
            2-3 готовых действия.
          </p>
        </Reveal>

        <Reveal delay={0.2}>
          <div className="mt-10 grid grid-cols-1 md:grid-cols-3 gap-4">
            {[
              { sym: 'sec', label: 'Секунды', desc: 'От команды до объяснения — две-три секунды. Без копипастов в браузер.' },
              { sym: 'any', label: 'Любой вывод', desc: 'stack traces, status сервисов, конфиги, JSON, логи — всё что попадает в твой терминал.' },
              { sym: 'safe', label: 'Без утечек', desc: 'Перед отправкой regex-фильтр чистит токены, JWT, пароли, email и абсолютные пути.' },
            ].map((v) => (
              <div
                key={v.sym}
                className="rounded-2xl border border-zinc-800/80 bg-zinc-900/40 p-5 text-left"
              >
                <div className="flex items-center gap-3 mb-2">
                  <span className="text-xs font-mono text-amber-400/80 uppercase tracking-widest">
                    {v.sym}
                  </span>
                  <span className="text-sm font-semibold text-zinc-100">{v.label}</span>
                </div>
                <p className="text-sm text-zinc-400 leading-relaxed">{v.desc}</p>
              </div>
            ))}
          </div>
        </Reveal>
      </div>
    </section>
  );
}

function LogosStrip() {
  const items = ['macOS', 'Linux', 'Windows', 'bash', 'zsh', 'fish', 'PowerShell'];
  return (
    <section className="bg-zinc-950">
      <div className="max-w-6xl mx-auto px-5 py-8">
        <p className="text-center text-xs uppercase tracking-widest text-zinc-600 mb-5">
          Работает везде
        </p>
        <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-4 opacity-70">
          {items.map((x) => (
            <span key={x} className="text-sm font-medium text-zinc-500">{x}</span>
          ))}
        </div>
      </div>
    </section>
  );
}

function Features() {
  const features = [
    {
      icon: Brain,
      title: 'Три провайдера на выбор',
      description: 'Claude (Anthropic), GPT-4o (OpenAI), Gemini (Google). Переключение одной командой. Твой ключ, твой счёт — мы ничего не проксируем.',
    },
    {
      icon: Terminal,
      title: 'Любой shell',
      description: 'bash, zsh, fish, PowerShell. wtf init ставит хук — после этого захват последнего вывода идёт автоматически.',
    },
    {
      icon: Zap,
      title: 'Спиннер и красивый вывод',
      description: 'Ответ рендерится прямо в терминале как Markdown — заголовки, инлайн-код, выделенные команды для копирования.',
    },
    {
      icon: ShieldCheck,
      title: 'Чистка секретов',
      description: '13 regex-правил: токены sk-/ghp_/AIza/xox*, JWT, Bearer, AWS-ключи, password=…, basic-auth URL, email. $HOME → ~. На первом запуске — явный consent-баннер.',
    },
    {
      icon: Layers,
      title: 'Локальный кеш',
      description: 'SHA-256 по нормализованному выводу + provider + язык. Тот же вывод второй раз — мгновенный ответ без обращения к API.',
    },
    {
      icon: Lock,
      title: 'Zero trust',
      description: 'Ни бэкенда, ни аккаунтов, ни телеметрии. Конфиг и кеш в ~/.wtf/ (mode 0600). Только исходящие HTTPS на API провайдера.',
    },
    {
      icon: Cpu,
      title: 'Один бинарь',
      description: 'Go, ~8 МБ, без зависимостей. brew, curl | sh, или скачай release для своей платформы.',
    },
    {
      icon: Globe,
      title: 'RU + EN',
      description: 'Объяснение на русском по умолчанию. wtf --lang en — на английском. Промпт оптимизирован под краткие ответы без воды.',
    },
    {
      icon: KeyRound,
      title: 'Свой ключ AI',
      description: 'wtf config — интерактивный wizard. Ключи лежат локально в ~/.wtf/config.json. Можно через env: ANTHROPIC_API_KEY / OPENAI_API_KEY / GEMINI_API_KEY.',
    },
  ];

  return (
    <section id="features" className="relative py-24 md:py-32 border-t border-zinc-900/80">
      <div className="max-w-6xl mx-auto px-5">
        <Reveal>
          <div className="max-w-2xl mx-auto text-center">
            <SectionDivider symbol="//" label="Capabilities" />
            <h2 className="text-3xl md:text-4xl font-semibold tracking-tight">
              Всё, что нужно для быстрого фикса
            </h2>
            <p className="mt-4 text-zinc-400 leading-relaxed">
              Утилита маленькая, но не игрушечная: продумано всё — от перехвата вывода
              до защиты твоих секретов перед отправкой в AI.
            </p>
          </div>
        </Reveal>

        <div className="mt-14 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {features.map((f, i) => (
            <FeatureCard key={f.title} {...f} delay={i * 0.05} />
          ))}
        </div>
      </div>
    </section>
  );
}

function HowItWorks() {
  const steps = [
    {
      num: '01',
      title: 'Установи',
      description: 'curl | bash или brew. Один бинарь, без зависимостей. Под macOS, Linux, Windows (PowerShell).',
    },
    {
      num: '02',
      title: 'Настрой провайдера',
      description: 'wtf config — wizard с тремя секциями (Claude / OpenAI / Gemini). Введи ключ хотя бы для одного. Ключи лежат локально в ~/.wtf/config.json.',
    },
    {
      num: '03',
      title: 'Поставь shell-хук',
      description: 'wtf init — добавит небольшой preexec/precmd-хук в .zshrc / .bashrc / fish config / PowerShell profile. Перезапусти shell.',
    },
    {
      num: '04',
      title: 'Упади и пиши wtf',
      description: 'После любой упавшей команды просто wtf — за пару секунд получишь объяснение и 2-3 варианта фикса.',
    },
  ];

  return (
    <section id="how" className="relative py-24 md:py-32 border-t border-zinc-900/80">
      <div className="max-w-6xl mx-auto px-5">
        <Reveal>
          <div className="text-center max-w-2xl mx-auto">
            <SectionDivider symbol=">_" label="The Flow" />
            <h2 className="text-3xl md:text-4xl font-semibold tracking-tight">
              Четыре шага до первого фикса
            </h2>
          </div>
        </Reveal>

        <div className="mt-14 grid grid-cols-1 md:grid-cols-4 gap-5 md:gap-8 relative">
          <div className="hidden md:block absolute top-9 left-[12%] right-[12%] h-px bg-gradient-to-r from-transparent via-zinc-800 to-transparent" />
          {steps.map((s, i) => (
            <Reveal key={s.num} delay={i * 0.06}>
              <div className="relative">
                <div className="w-[72px] h-[72px] rounded-2xl border border-zinc-800 bg-zinc-900/60 backdrop-blur flex items-center justify-center mb-5 mx-auto">
                  <span className="text-2xl font-mono font-semibold text-zinc-300 tracking-tight">
                    {s.num}
                  </span>
                </div>
                <h3 className="text-lg font-semibold text-center text-zinc-100 mb-2">{s.title}</h3>
                <p className="text-sm text-zinc-400 leading-relaxed text-center max-w-xs mx-auto">
                  {s.description}
                </p>
              </div>
            </Reveal>
          ))}
        </div>

        <Reveal delay={0.2}>
          <div id="install" className="mt-14 max-w-2xl mx-auto">
            <InstallCommand />
            <p className="mt-3 text-center text-xs text-zinc-500">
              После установки доступны команды:{' '}
              <code className="text-zinc-400">wtf · wtf init · wtf config · wtf --rerun · wtf --explain "..."</code>
            </p>
          </div>
        </Reveal>
      </div>
    </section>
  );
}

function DemoSection() {
  return (
    <section className="relative py-24 md:py-32 border-t border-zinc-900/80 overflow-hidden">
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[400px] bg-amber-400/5 blur-[120px] rounded-full" />
      </div>

      <div className="relative max-w-3xl mx-auto px-5 text-center">
        <Reveal>
          <SectionDivider symbol="$_" label="The Output" />
          <h2 className="text-3xl md:text-4xl font-semibold tracking-tight leading-tight">
            Как выглядит ответ
          </h2>
          <p className="mt-4 text-zinc-400 leading-relaxed">
            Одно предложение про то, что произошло или что показывает вывод. Если есть что
            делать — 2-3 готовые команды для копирования. Опционально — короткое «почему»
            и одна ссылка на доку, если она реально есть. Без воды и переливания из пустого.
          </p>
          <p className="mt-3 text-zinc-500 leading-relaxed text-sm">
            На повторных одинаковых запросах срабатывает локальный кеш и ответ приходит мгновенно.
          </p>
        </Reveal>
      </div>
    </section>
  );
}

function FAQ() {
  const items = [
    {
      q: 'Как wtf вообще читает, что у меня было в терминале?',
      a: 'Через shell-хук: wtf init добавляет небольшой preexec/precmd-хук в .zshrc / .bashrc / fish config / PowerShell profile. Хук пишет последнюю команду и её exit-код в ~/.wtf/last_meta. При вызове wtf утилита читает этот файл и отправляет вывод в AI — независимо от того, упало оно или просто хочется разобраться. Если хука нет или вывод пустой — wtf может перезапустить последнюю команду (--rerun) и поймать stdout/stderr. Также можно явно: wtf --explain "<любой текст>".',
    },
    {
      q: 'Это работает только с ошибками?',
      a: 'Нет. wtf берёт последний вывод терминала независимо от exit-кода. Хочешь разобраться, что показывает systemctl status nginx или что внутри длинного JSON от curl — пиши wtf, объяснит. Для упавших команд дополнительно идут 2-3 варианта починить.',
    },
    {
      q: 'Это правда полностью open-source? Никаких подписок?',
      a: 'Да. Один Go-бинарь, MIT-лицензия. Нет своего бэкенда, нет аккаунтов, нет телеметрии. Запросы идут напрямую в API провайдера на твой ключ. Платишь только за API-вызовы.',
    },
    {
      q: 'Какой провайдер лучше выбрать?',
      a: 'По умолчанию — Claude Haiku 4.5: быстро, дёшево, качественно для коротких объяснений. OpenAI gpt-4o-mini — сравним. Gemini 2.0 Flash — самый дешёвый и быстрый. wtf config позволяет переключаться, wtf --provider <name> — разово.',
    },
    {
      q: 'А мои секреты не утекут? Там же могут быть токены в логах.',
      a: 'Перед отправкой работает 13 regex-правил: sk-ant-…, sk-…, AIza…, gh*_…, xox*-, JWT (eyJ…), Bearer …, AKIA…, aws_secret=…, private keys (-----BEGIN…), password/secret/token=…, basic-auth URL (user:pass@host), email. Плюс $HOME → ~. На первом запуске показывается баннер с тем, что именно уйдёт в API, и это сохраняется в конфиге как «согласие получено».',
    },
    {
      q: 'Что если я хочу видеть, что отправляется?',
      a: 'wtf при первом вызове показывает консент-баннер с полным списком метаданных (OS, shell, cwd, git branch, package manager, команда, exit code, размер вывода, какие классы секретов вычистились). Дальше можно посмотреть последний отправленный запрос в ~/.wtf/cache/<hash>.json — там лежит сохранённый ответ.',
    },
    {
      q: 'Безопасно ли запускать curl | sudo bash?',
      a: 'Скрипт короткий, читай его перед запуском: github.com/kitay-sudo/wtf/blob/main/install.sh. Он только определяет архитектуру, скачивает бинарь из GitHub Releases и кладёт в /usr/local/bin/wtf. Никаких внешних серверов кроме github.com.',
    },
    {
      q: 'Можно ли работать офлайн?',
      a: 'Не полностью. AI-провайдеры все в облаке. Но ответы кешируются по хешу вывода в ~/.wtf/cache/, так что повторные одинаковые запросы возвращаются мгновенно из кеша без сети. Self-hosted режим через ollama в roadmap.',
    },
    {
      q: 'Сколько это в RAM/диск?',
      a: 'Бинарь ~8 МБ. RAM при вызове <20 МБ, между вызовами — 0 (это не демон, это разовый запуск). Конфиг ~1 КБ, кеш растёт по 2-5 КБ на ответ, чистится автоматически через 30 дней.',
    },
    {
      q: 'А если я в Windows?',
      a: 'Работает через PowerShell (Windows PowerShell 5.1 и pwsh 7+). wtf init установит хук в $PROFILE. Поддерживается также через Git Bash / WSL — там работает unix-хук.',
    },
  ];

  return (
    <section id="faq" className="py-24 md:py-32 border-t border-zinc-900/80">
      <div className="max-w-3xl mx-auto px-5">
        <Reveal>
          <div className="text-center">
            <SectionDivider symbol="??" label="Questions" />
            <h2 className="text-3xl md:text-4xl font-semibold tracking-tight">
              Ответы на самое важное
            </h2>
          </div>
        </Reveal>

        <Reveal delay={0.1}>
          <div className="mt-10">
            {items.map((it) => (
              <FAQItem key={it.q} question={it.q} answer={it.a} />
            ))}
          </div>
        </Reveal>
      </div>
    </section>
  );
}

function CTA() {
  return (
    <section className="relative py-24 md:py-32 border-t border-zinc-900/80 overflow-hidden">
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[400px] bg-amber-400/10 blur-[120px] rounded-full" />
      </div>
      <div className="relative max-w-3xl mx-auto px-5 text-center">
        <Reveal>
          <div className="inline-flex items-center justify-center mb-6">
            <WtfMark size={56} />
          </div>
          <h2 className="text-3xl md:text-5xl font-semibold tracking-tight leading-tight">
            Перестань копировать вывод в Google.<br />
            <span className="bg-gradient-to-r from-amber-400 to-yellow-300 bg-clip-text text-transparent">
              Просто пиши wtf.
            </span>
          </h2>
          <p className="mt-5 text-zinc-400 max-w-xl mx-auto">
            60 секунд от установки до первого объяснения. Один бинарь, твой ключ, никаких подписок.
          </p>

          <div className="mt-8 max-w-2xl mx-auto">
            <InstallCommand />
          </div>

          <div className="mt-5 flex items-center justify-center">
            <a
              href={REPO_URL}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-2 text-zinc-300 hover:text-zinc-100 border border-zinc-800 hover:border-zinc-700 rounded-xl px-5 py-3 transition-colors"
            >
              <Github size={16} />
              Посмотреть код на GitHub
            </a>
          </div>
        </Reveal>
      </div>
    </section>
  );
}

const CHANGELOG = [
  {
    version: '0.1.0',
    date: '2026-04-30',
    title: 'Первый публичный релиз',
    highlights: [
      'CLI на Go: один кросс-платформенный бинарник',
      'Три провайдера на выбор: Claude, OpenAI, Gemini',
      'Shell-хук для bash / zsh / fish / PowerShell + fallback через --rerun',
      'Локальный кеш с TTL 30 дней',
      'Чистка секретов: 13 regex-правил, consent-баннер при первом запуске',
      'Spinner в стиле Claude CLI, цветной Markdown-рендер ответа',
    ],
  },
];

const RELEASES_URL = 'https://github.com/kitay-sudo/wtf/releases';
const CHANGELOG_URL = 'https://github.com/kitay-sudo/wtf/blob/main/CHANGELOG.md';

function Changelog() {
  const dateFormatter = new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });

  return (
    <section id="changelog" className="relative py-24 md:py-32 border-t border-zinc-900/80 overflow-hidden">
      <div className="relative max-w-3xl mx-auto px-5">
        <Reveal>
          <div className="text-center">
            <SectionDivider symbol="++" label="The Chronicle" />
            <h2 className="text-3xl md:text-4xl font-semibold tracking-tight">
              Журнал изменений
            </h2>
            <p className="mt-4 text-zinc-400 leading-relaxed max-w-xl mx-auto">
              Что меняется в каждом релизе и когда он был выпущен — чтобы было видно,
              что проект живой.
            </p>
          </div>
        </Reveal>

        <Reveal delay={0.1}>
          <ol className="mt-12 relative border-l border-zinc-800/80 ml-3">
            {CHANGELOG.map((rel, idx) => {
              const isLatest = idx === 0;
              const dateLabel = (() => {
                const d = new Date(rel.date);
                return Number.isNaN(d.getTime()) ? rel.date : dateFormatter.format(d);
              })();

              return (
                <li key={rel.version} className="relative pl-8 pb-10 last:pb-0">
                  <span
                    className={`absolute -left-[7px] top-1.5 w-3.5 h-3.5 rounded-full border-2 ${
                      isLatest
                        ? 'border-amber-400 bg-amber-400/30 shadow-[0_0_0_4px_rgba(251,191,36,0.08)]'
                        : 'border-zinc-700 bg-zinc-900'
                    }`}
                    aria-hidden
                  />

                  <div className="flex flex-wrap items-center gap-2 mb-1">
                    <span className="text-base md:text-lg font-mono font-semibold text-zinc-100">
                      v{rel.version}
                    </span>
                    {isLatest && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full border border-amber-400/40 bg-amber-400/10 text-amber-300 text-[10px] uppercase tracking-widest font-semibold">
                        <Sparkles size={10} />
                        Latest
                      </span>
                    )}
                    <span className="text-xs text-zinc-500 font-mono ml-auto">
                      <time dateTime={rel.date}>{dateLabel}</time>
                    </span>
                  </div>

                  <h3 className="text-sm md:text-base font-semibold text-zinc-200 mb-3">
                    {rel.title}
                  </h3>

                  <ul className="space-y-1.5">
                    {rel.highlights.map((h, i) => (
                      <li
                        key={i}
                        className="flex gap-2 text-sm text-zinc-400 leading-relaxed"
                      >
                        <span className="shrink-0 mt-2 w-1 h-1 rounded-full bg-zinc-600" />
                        <span>{h}</span>
                      </li>
                    ))}
                  </ul>
                </li>
              );
            })}
          </ol>
        </Reveal>

        <Reveal delay={0.2}>
          <div className="mt-10 flex flex-wrap items-center justify-center gap-3">
            <a
              href={CHANGELOG_URL}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-2 text-sm text-zinc-300 hover:text-zinc-100 border border-zinc-800 hover:border-zinc-700 rounded-lg px-4 py-2 transition-colors"
            >
              <History size={14} />
              Полный CHANGELOG
            </a>
            <a
              href={RELEASES_URL}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-2 text-sm text-zinc-300 hover:text-zinc-100 border border-zinc-800 hover:border-zinc-700 rounded-lg px-4 py-2 transition-colors"
            >
              <Github size={14} />
              Все релизы на GitHub
            </a>
          </div>
        </Reveal>
      </div>
    </section>
  );
}

// Стена донатеров. Чтобы добавить нового — допиши объект в массив и сделай commit.
// Поля:
//   handle    — ник (с @ или без, при отрисовке @ всё равно срезается)
//   amount    — строка для отображения, например "0.66 TON" или "10 USDT"
//   amountTon — число в TON для сортировки (используется ТОЛЬКО для определения top)
//   addedAt   — дата в ISO ("2026-04-30"), нужна для роли first-ever (первая десятка)
//   note      — опционально, короткая ремарка
//
// Роли расставляются автоматически:
//   • TOP DONOR  — у кого amountTon максимален (золотой акцент)
//   • FIRST EVER — первые 10 по addedAt (серебряный шильдик, не отбирается)
const DONORS = [];

const TELEGRAM_HANDLE = '@kitay9';
const TELEGRAM_URL = 'https://t.me/kitay9';

function Support() {
  const wallets = [
    {
      label: 'USDT',
      network: 'TRON · TRC20',
      address: 'TF9F2FPkreHVfbe8tZtn4V76j3jLo4SeXM',
    },
    {
      label: 'TON',
      network: 'The Open Network',
      address: 'UQBl88kXWJWyHkDPkWNYQwwSCiCAIfA2DiExtZElwJFlIc1o',
    },
  ];

  return (
    <section id="support" className="relative py-24 md:py-32 border-t border-zinc-900/80 overflow-hidden">
      <div className="relative max-w-3xl mx-auto px-5">
        <Reveal>
          <div className="text-center">
            <SectionDivider symbol="<3" label="Gratitude" />
            <h2 className="text-3xl md:text-4xl font-semibold tracking-tight">
              Поддержать проект
            </h2>
            <p className="mt-4 text-zinc-400 leading-relaxed max-w-xl mx-auto">
              wtf пилится в свободное время, без подписок и платных тарифов. Если он
              сэкономил тебе пару часов — поддержать можно тремя способами ниже.
              Деньги идут на новые провайдеры, ускорение релизов и время на доведение
              фич из roadmap.
            </p>
            <p className="mt-3 text-zinc-500 text-sm leading-relaxed max-w-xl mx-auto">
              Поддержавшие попадают в{' '}
              <a href="#support" className="text-amber-300/90 hover:text-amber-200 underline-offset-4 hover:underline">
                стену чести
              </a>{' '}
              ниже — публичный список тех, кто помог проекту встать на ноги.
            </p>
          </div>
        </Reveal>

        <Reveal delay={0.08}>
          <TimewebCard />
        </Reveal>

        <Reveal delay={0.12}>
          <div className="mt-6 text-center">
            <p className="text-xs uppercase tracking-widest text-zinc-500 font-mono">
              или криптой напрямую
            </p>
          </div>
        </Reveal>

        <Reveal delay={0.15}>
          <div className="mt-6 grid grid-cols-1 md:grid-cols-2 gap-4">
            {wallets.map((w, i) => (
              <WalletCard key={w.label} {...w} delay={i * 0.05} />
            ))}
          </div>
        </Reveal>

        <Reveal delay={0.15}>
          <div className="mt-10 rounded-2xl border border-amber-400/20 bg-gradient-to-br from-amber-400/5 via-zinc-900/40 to-zinc-900/40 p-6 md:p-8">
            <div className="flex items-start gap-4">
              <div className="shrink-0 inline-flex items-center justify-center w-11 h-11 rounded-xl border border-amber-400/30 bg-amber-400/10 text-amber-400">
                <Send size={18} />
              </div>
              <div className="flex-1 min-w-0">
                <h3 className="text-base md:text-lg font-semibold text-zinc-100">
                  Хочешь попасть в стену чести?
                </h3>
                <p className="mt-1.5 text-sm text-zinc-400 leading-relaxed">
                  После доната напиши в Telegram{' '}
                  <a
                    href={TELEGRAM_URL}
                    target="_blank"
                    rel="noreferrer"
                    className="text-amber-400 hover:text-amber-300 font-mono"
                  >
                    {TELEGRAM_HANDLE}
                  </a>{' '}
                  свой ник — добавлю в список ниже навсегда.
                </p>
                <a
                  href={TELEGRAM_URL}
                  target="_blank"
                  rel="noreferrer"
                  className="mt-4 inline-flex items-center gap-1.5 text-xs font-medium px-3 py-2 rounded-md border border-amber-400/30 hover:border-amber-400/60 hover:bg-amber-400/10 transition-colors text-amber-300"
                >
                  <Send size={13} />
                  Написать в Telegram
                </a>
              </div>
            </div>
          </div>
        </Reveal>

        <Reveal delay={0.2}>
          <DonorsWall donors={DONORS} />
        </Reveal>
      </div>
    </section>
  );
}

function DonorsWall({ donors }) {
  const empty = !donors || donors.length === 0;

  // Особые роли: top (по сумме) и first-ever (по дате).
  // Один и тот же донатер может держать обе роли — тогда показываем одну
  // карточку с двумя плашками. Когда придёт следующий с большей суммой —
  // он станет top, а старый top-first останется только first-ever.
  const annotated = !empty ? annotateDonors(donors) : [];
  const honors = annotated.filter((d) => d.roles.length > 0);
  const rest = annotated.filter((d) => d.roles.length === 0);

  return (
    <div className="mt-10">
      <div className="flex items-center gap-3 mb-2">
        <Heart size={14} className="text-amber-400" strokeWidth={2.4} />
        <h3 className="text-sm font-semibold tracking-wide uppercase text-zinc-300">
          Стена чести
        </h3>
        {!empty && (
          <span className="text-xs text-zinc-500 font-mono ml-auto">
            {donors.length} {donors.length === 1 ? 'человек' : 'человек'}
          </span>
        )}
      </div>
      {!empty && (
        <p className="mb-5 text-xs text-zinc-500 leading-relaxed">
          <span className="text-amber-300/90 font-mono">top</span> — крупнейший донат на сейчас.{' '}
          <span className="text-zinc-300 font-mono">first ever</span> — первая десятка тех, кто
          поддержал проект раньше всех. Этот шильдик не отбирается и достаётся только им —
          навсегда.
        </p>
      )}

      {empty ? (
        <div className="rounded-xl border border-dashed border-zinc-800 bg-zinc-900/30 p-8 text-center">
          <p className="text-sm text-zinc-500">
            Пока пусто.{' '}
            <a
              href={TELEGRAM_URL}
              target="_blank"
              rel="noreferrer"
              className="text-amber-400 hover:text-amber-300 font-medium"
            >
              Будь первым
            </a>{' '}
            — твой ник окажется здесь и останется навсегда.
          </p>
        </div>
      ) : (
        <div className="space-y-5">
          {honors.length > 0 && (
            <div
              className={
                honors.length === 1
                  ? 'grid grid-cols-1 gap-4'
                  : 'grid grid-cols-1 md:grid-cols-2 gap-4'
              }
            >
              {honors.map((d) => (
                <HonorCard key={d.handle} donor={d} />
              ))}
            </div>
          )}

          {rest.length > 0 && (
            <ul className="flex flex-wrap gap-2">
              {rest.map((d) => {
                const nick = d.handle.replace(/^@/, '');
                return (
                  <li
                    key={nick}
                    title={d.note || nick}
                    className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full border border-zinc-800 bg-zinc-900/50 text-sm text-zinc-300 font-mono"
                  >
                    <Heart size={11} className="text-amber-400/70" fill="currentColor" />
                    {nick}
                    {d.amount && (
                      <span className="text-zinc-500 text-[11px]">· {d.amount}</span>
                    )}
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}

const FIRST_EVER_LIMIT = 10;

function annotateDonors(donors) {
  let topIdx = -1;
  let topAmount = -Infinity;
  donors.forEach((d, i) => {
    const a = typeof d.amountTon === 'number' ? d.amountTon : -Infinity;
    if (a > topAmount) {
      topAmount = a;
      topIdx = i;
    }
  });

  const firstIdxs = new Set(
    donors
      .map((d, i) => ({ i, t: d.addedAt ? new Date(d.addedAt).getTime() : NaN }))
      .filter((x) => !Number.isNaN(x.t))
      .sort((a, b) => a.t - b.t)
      .slice(0, FIRST_EVER_LIMIT)
      .map((x) => x.i),
  );

  return donors.map((d, i) => {
    const roles = [];
    if (i === topIdx && topAmount > -Infinity) roles.push('top');
    if (firstIdxs.has(i)) roles.push('first-ever');
    return { ...d, roles };
  });
}

function HonorCard({ donor }) {
  const nick = donor.handle.replace(/^@/, '');
  const isTop = donor.roles.includes('top');
  const isFirst = donor.roles.includes('first-ever');

  // Top → золотая, иначе (только first-ever) → серебряная.
  const gold = isTop;
  const palette = gold
    ? {
        border: 'border-amber-400/40',
        bg: 'from-amber-400/10 via-zinc-900/60 to-zinc-900/40',
        iconWrap: 'border-amber-400/50 bg-amber-400/10 text-amber-300',
      }
    : {
        border: 'border-zinc-400/25',
        bg: 'from-zinc-300/5 via-zinc-900/60 to-zinc-900/40',
        iconWrap: 'border-zinc-400/40 bg-zinc-300/10 text-zinc-200',
      };

  return (
    <div
      title={donor.note || nick}
      className={`relative block overflow-hidden rounded-2xl border bg-gradient-to-br p-5 md:p-6 ${palette.border} ${palette.bg}`}
    >
      <div className="relative flex items-start gap-4">
        <div
          className={`shrink-0 inline-flex items-center justify-center w-12 h-12 rounded-xl border ${palette.iconWrap}`}
        >
          <Heart size={20} fill="currentColor" />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            {isTop && (
              <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full border border-amber-400/40 bg-amber-400/10 text-amber-300 text-[10px] uppercase tracking-widest font-semibold font-mono">
                ★ top
              </span>
            )}
            {isFirst && (
              <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full border border-zinc-400/40 bg-zinc-300/10 text-zinc-200 text-[10px] uppercase tracking-widest font-semibold font-mono">
                01 first ever
              </span>
            )}
          </div>

          <div className="mt-2 flex flex-wrap items-baseline gap-x-3 gap-y-1">
            <span className="text-xl md:text-2xl font-semibold font-mono text-zinc-100">
              {nick}
            </span>
            {donor.amount && (
              <span className="text-sm font-mono text-amber-300">{donor.amount}</span>
            )}
          </div>

          {(donor.note || (isTop && isFirst)) && (
            <p className="mt-1.5 text-xs text-zinc-500 leading-relaxed">
              {isTop && isFirst
                ? donor.note || 'и первый по времени, и пока крупнейший донат'
                : isTop
                  ? donor.note || 'крупнейший донат на момент сейчас'
                  : donor.note || 'один из первой десятки поддержавших'}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}

// TimewebCard — отдельный способ поддержать проект, который НИЧЕГО не стоит юзеру:
// он регистрируется по нашей реф-ссылке (получает свой сервер),
// мы получаем небольшую партнёрскую комиссию.
//
// Disclosure прямо в карточке: "партнёрская ссылка". Никаких маскировок —
// эта секция и так помечена как способ поддержки, юзер понимает контекст.
function TimewebCard() {
  return (
    <div className="mt-10 rounded-2xl border border-amber-400/30 bg-gradient-to-br from-amber-400/10 via-zinc-900/60 to-zinc-900/40 p-6 md:p-7">
      <div className="flex items-start gap-4">
        <div className="shrink-0 inline-flex items-center justify-center w-12 h-12 rounded-xl border border-amber-400/40 bg-amber-400/10 text-amber-300">
          <Server size={20} />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex flex-wrap items-center gap-2 mb-1.5">
            <h3 className="text-base md:text-lg font-semibold text-zinc-100">
              Поднять сервер на Timeweb Cloud
            </h3>
            <span className="inline-flex items-center px-1.5 py-0.5 rounded border border-zinc-700 bg-zinc-900/60 text-[10px] uppercase tracking-widest font-mono text-zinc-400">
              партнёрская ссылка
            </span>
          </div>
          <p className="text-sm text-zinc-400 leading-relaxed">
            Самый честный способ поддержать проект — без расходов из своего кармана.
            Регистрируешься на{' '}
            <a
              href="https://timeweb.cloud/?i=104289"
              target="_blank"
              rel="sponsored noopener"
              className="text-amber-300 hover:text-amber-200 font-medium underline-offset-4 hover:underline"
            >
              Timeweb Cloud
            </a>{' '}
            по нашей ссылке, получаешь VPS под свои задачи — и нам тоже капает
            небольшая комиссия. В качестве бонуса —{' '}
            <span className="text-zinc-200">помощь с настройкой сервера</span>:
            напиши в Telegram{' '}
            <a
              href="https://t.me/kitay9"
              target="_blank"
              rel="noreferrer"
              className="text-amber-300 hover:text-amber-200 font-mono"
            >
              @kitay9
            </a>{' '}
            — подскажу с конфигом, файрволом, деплоем своих сервисов.
          </p>
          <a
            href="https://timeweb.cloud/?i=104289"
            target="_blank"
            rel="sponsored noopener"
            className="mt-4 inline-flex items-center gap-1.5 text-sm font-medium px-4 py-2 rounded-lg bg-amber-400 hover:bg-amber-300 text-zinc-950 transition-colors"
          >
            Перейти к Timeweb Cloud
            <ArrowRight size={14} />
          </a>
        </div>
      </div>
    </div>
  );
}

function WalletCard({ label, network, address }) {
  const [copied, setCopied] = useState(false);
  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(address);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      /* empty */
    }
  };

  return (
    <div className="rounded-2xl border border-zinc-800 bg-zinc-900/40 p-5 backdrop-blur">
      <div className="flex items-baseline justify-between mb-3">
        <span className="text-base font-semibold text-zinc-100">{label}</span>
        <span className="text-xs text-zinc-500 font-mono">{network}</span>
      </div>
      <div className="flex items-center gap-2 rounded-lg border border-zinc-800 bg-zinc-950/60 p-3">
        <code className="text-xs text-amber-300 font-mono break-all flex-1 min-w-0">
          {address}
        </code>
        <button
          onClick={onCopy}
          className="shrink-0 inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1.5 rounded-md border border-zinc-700 hover:border-amber-400/50 hover:bg-amber-400/10 transition-colors text-zinc-300"
          aria-label={`Скопировать адрес ${label}`}
        >
          {copied ? <Check size={14} className="text-amber-400" /> : <Copy size={14} />}
          {copied ? 'Скопировано' : 'Копировать'}
        </button>
      </div>
    </div>
  );
}

function Footer() {
  return (
    <footer className="border-t border-zinc-900/80 py-10">
      <div className="max-w-6xl mx-auto px-5 flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-2 text-sm text-zinc-500">
          <WtfMark size={18} />
          <span>wtf · MIT · © {new Date().getFullYear()}</span>
          <span className="text-zinc-700">·</span>
          <a
            href="https://github.com/kitay-sudo"
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-center gap-1 text-zinc-500 hover:text-amber-400 transition-colors"
          >
            by <Github size={12} /> kitay-sudo
          </a>
        </div>
        <div className="flex items-center gap-5 text-sm text-zinc-500">
          <a href="#features" className="hover:text-zinc-300 transition-colors">Возможности</a>
          <a href="#install" className="hover:text-zinc-300 transition-colors">Установка</a>
          <a href="#faq" className="hover:text-zinc-300 transition-colors">FAQ</a>
          <a href="#changelog" className="hover:text-zinc-300 transition-colors">Изменения</a>
          <a href="#support" className="hover:text-zinc-300 transition-colors">Поддержать</a>
          <a href={REPO_URL} target="_blank" rel="noreferrer" className="hover:text-zinc-300 transition-colors flex items-center gap-1.5">
            <Github size={14} /> GitHub
          </a>
        </div>
      </div>
    </footer>
  );
}
