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
            Опиши проблему словами — <br />
            <span className="inline-flex items-center gap-3 align-middle">
              <span className="bg-gradient-to-r from-amber-400 to-yellow-300 bg-clip-text text-transparent">
                wtf
              </span>
              <WtfMark size={56} />
            </span>{' '}
            <span className="text-zinc-200">сам всё проверит</span>
          </h1>
        </Reveal>

        <Reveal delay={0.08}>
          <p className="mt-6 text-center text-lg md:text-xl text-zinc-200 max-w-2xl mx-auto leading-relaxed font-medium">
            Терминальный sysadmin-агент. Пиши <code className="font-mono text-amber-300">wtf nginx не стартует</code> — он сам выполнит диагностические команды, прочитает логи, найдёт причину и покажет точные команды для починки.
          </p>
        </Reveal>

        <Reveal delay={0.16}>
          <p className="mt-5 text-center text-sm md:text-base text-zinc-400 max-w-2xl mx-auto leading-relaxed">
            Один Go-бинарник. Три AI-провайдера: Claude, OpenAI, Gemini — твой ключ, твой счёт.
            Запоминает что узнал о сервере между сессиями. Никогда не запускает sudo без твоего ведома. Open-source, MIT.
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
            Бесплатно · MIT-лицензия · macOS / Linux / Windows · никаких хуков и обёрток
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
            Это первое, что ты пишешь, когда сервер начинает странно себя вести —
            nginx не стартует, диск заполнен непонятно чем, что-то жрёт CPU.
            Раньше — копировал ошибки в ChatGPT, гуглил команды, выполнял по одной.
            Теперь — пишешь <code className="font-mono text-amber-300">wtf nginx не стартует</code>{' '}
            и смотришь как агент сам всё разруливает.
          </p>
        </Reveal>

        <Reveal delay={0.15}>
          <p className="mt-4 text-zinc-400 leading-relaxed md:text-lg">
            Агент видит твой вопрос, выполняет на машине безопасные read-only команды
            (<code className="font-mono text-amber-300">systemctl status</code>,{' '}
            <code className="font-mono text-amber-300">journalctl</code>,{' '}
            <code className="font-mono text-amber-300">cat /etc/...</code>),
            читает их вывод, итеративно докапывается до причины. Когда нужно что-то поменять —
            показывает точную команду чтобы ты выполнил сам. <span className="text-amber-300">sudo никогда не запускается без твоего ведома.</span>
          </p>
        </Reveal>

        <Reveal delay={0.2}>
          <div className="mt-10 grid grid-cols-1 md:grid-cols-3 gap-4">
            {[
              { sym: 'agent', label: 'Сам диагностирует', desc: 'До 15 раундов: запускает команды, читает выводы, делает выводы. Юзер только смотрит.' },
              { sym: 'memo', label: 'Помнит контекст', desc: 'Между сессиями запоминает что узнал о машине — версии, пути, особенности. Не диагностирует с нуля каждый раз.' },
              { sym: 'safe', label: 'Безопасный', desc: 'Whitelist read-only команд для авто-запуска. sudo, rm, restart — только показывает, не делает сам.' },
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
      title: 'Tool-use агент',
      description: 'Модель сама решает какие команды запустить, читает их вывод, итерирует. До 15 раундов диагностики на сессию. Tool-use API всех трёх провайдеров — никакого хрупкого парсинга текста.',
    },
    {
      icon: ShieldCheck,
      title: 'Safe-by-default',
      description: 'Встроенный classifier разделяет команды на read-only (выполняем сами) и destructive (показываем юзеру). sudo, rm, restart, install никогда не запускаются автоматически.',
    },
    {
      icon: History,
      title: 'Память между сессиями',
      description: 'После каждой проблемы агент сохраняет ключевые факты: версии сервисов, пути конфигов, найденные решения. В следующий раз уже знает контекст — не диагностирует с нуля.',
    },
    {
      icon: Cpu,
      title: 'Три провайдера',
      description: 'Claude (Anthropic), OpenAI, Gemini. Все три используют tool-use API. Переключение одной командой. Твой ключ, твой счёт — ничего не проксируется.',
    },
    {
      icon: Lock,
      title: 'Чистка секретов',
      description: '13 regex-правил перед отправкой в API и перед записью в память: токены sk-/ghp_/AIza/xox*, JWT, Bearer, AWS-ключи, password=, basic-auth URL, email. $HOME → ~.',
    },
    {
      icon: Zap,
      title: 'Quiet UI с timestamps',
      description: 'Каждая строка с временем [HH:MM:SS], спиннер во время выполнения, итог одной строкой. Полный вывод доступен через wtf -v. Auto-retry при rate-limit от провайдера.',
    },
    {
      icon: Terminal,
      title: 'Никаких хуков',
      description: 'Не правит .bashrc, не подменяет шелл, не оборачивает команды. Просто бинарь — пишешь wtf <вопрос>, получаешь ответ. Работает в любом терминале.',
    },
    {
      icon: Globe,
      title: 'RU + EN',
      description: 'Объяснение на русском по умолчанию. wtf --lang en — на английском. Промпт оптимизирован под краткие ответы без воды.',
    },
    {
      icon: KeyRound,
      title: 'Свой ключ AI',
      description: 'wtf config — интерактивный wizard. Ключи лежат локально в ~/.wtf/config.json (mode 0600). Можно через env: ANTHROPIC_API_KEY / OPENAI_API_KEY / GEMINI_API_KEY.',
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
      description: 'wtf config — wizard с Claude / OpenAI / Gemini. Введи ключ хотя бы для одного. Ключи лежат локально в ~/.wtf/config.json (mode 0600).',
    },
    {
      num: '03',
      title: 'Опиши проблему — wtf разруливает',
      description: 'wtf nginx не стартует. Агент сам выполнит диагностические команды, прочитает их вывод, найдёт причину и покажет решение. sudo показывает чтобы ты выполнил сам.',
    },
  ];

  return (
    <section id="how" className="relative py-24 md:py-32 border-t border-zinc-900/80">
      <div className="max-w-6xl mx-auto px-5">
        <Reveal>
          <div className="text-center max-w-2xl mx-auto">
            <SectionDivider symbol=">_" label="The Flow" />
            <h2 className="text-3xl md:text-4xl font-semibold tracking-tight">
              Три шага до работающего сервера
            </h2>
          </div>
        </Reveal>

        <div className="mt-14 grid grid-cols-1 md:grid-cols-3 gap-5 md:gap-8 relative">
          <div className="hidden md:block absolute top-9 left-[16%] right-[16%] h-px bg-gradient-to-r from-transparent via-zinc-800 to-transparent" />
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
              После установки:{' '}
              <code className="text-zinc-400">wtf &lt;вопрос&gt; · wtf config · wtf memory show · wtf -v</code>
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
            Как выглядит работа агента
          </h2>
          <p className="mt-4 text-zinc-400 leading-relaxed">
            Каждая строка с временем <code className="font-mono text-amber-300">[HH:MM:SS]</code>:
            спиннер во время выполнения команды, итог одной строкой после.
            Команды требующие <code className="font-mono text-amber-300">sudo</code>{' '}
            или меняющие систему помечаются <span className="text-amber-300">⚠</span> и
            показываются для ручного выполнения. В конце — ★ финальный ответ с решением.
          </p>
          <p className="mt-3 text-zinc-500 leading-relaxed text-sm">
            Полный вывод диагностических команд скрыт по умолчанию (флаг <code className="font-mono text-zinc-400">-v</code> для verbose).
            При ошибке команды — последние 5 строк показываются автоматически.
          </p>
        </Reveal>
      </div>
    </section>
  );
}

function FAQ() {
  const items = [
    {
      q: 'Как wtf понимает что у меня сломалось?',
      a: 'Ты словами описываешь проблему: wtf nginx не стартует. Агент собирает контекст (ОС, shell, cwd, package manager, git, плюс свою память о машине из прошлых сессий) и сам выполняет диагностические команды через tool-use API: systemctl status, journalctl, nginx -t, ls, cat — что нужно. Читает их вывод, итерирует, докапывается до причины. До 15 раундов на одну сессию. Никаких shell-хуков, обёрток, /tmp-файлов — обычный бинарь.',
    },
    {
      q: 'Какие команды агент может запускать сам?',
      a: 'Только read-only из встроенного whitelist: ls, cat, tail, grep, ps, df, ip, ss, lsof, systemctl status, journalctl, docker ps, docker logs, git status, git log, nginx -t, apt list, dpkg -l, и т.п. (~80 утилит). Любая destructive команда — sudo, rm, mv, dd, chmod, chown, restart, install, push, force, опасные паттерны (cmd | sh, > /etc/...) — блокируется и показывается юзеру для ручного выполнения. Никогда не запускается автоматически.',
    },
    {
      q: 'Что значит "память между сессиями"?',
      a: 'После каждой решённой проблемы агент сохраняет короткие заметки 4 типов: machine_fact (стабильные факты — ОС, версии), service_state (состояние сервисов, домены), user_preference (привычки), resolved_issue (что чинили). В ~/.wtf/memory/store.json. При следующем wtf эти заметки идут в system-промпт — агент сразу знает контекст. Раз в 20 сессий запускается консолидация: AI сжимает старые заметки в 30-50 самых ценных.',
    },
    {
      q: 'Это правда полностью open-source? Никаких подписок?',
      a: 'Да. Один Go-бинарь, MIT-лицензия. Нет своего бэкенда, нет аккаунтов, нет телеметрии. Запросы идут напрямую в API провайдера на твой ключ. Платишь только за API-вызовы.',
    },
    {
      q: 'Какой провайдер лучше выбрать?',
      a: 'Все три используют tool-use API. Claude Haiku 4.5 — быстро и дёшево. Sonnet 4.6 — медленнее но умнее, лучше для сложной диагностики. OpenAI gpt-4o-mini — сравним с Haiku. Gemini 2.0 Flash — самый дешёвый и с большими лимитами на токены/мин (плюс при долгих сессиях). wtf config позволяет переключаться, wtf --provider <name> — разово.',
    },
    {
      q: 'А мои секреты не утекут? Я работаю с чувствительными конфигами.',
      a: 'Перед отправкой в API и перед записью в память работают 13 regex-правил: sk-ant-, sk-, AIza, gh*_, xox*-, JWT (eyJ.eyJ.), Bearer, AKIA, aws_secret=, private keys, password/secret/token=, basic-auth URL, email. Плюс $HOME → ~. И в память пишется только то что AI явно пометил для запоминания (через notes в finish), не весь stdout команд.',
    },
    {
      q: 'А если упрусь в rate-limit провайдера?',
      a: 'Между раундами автопауза 800мс — снижает шанс схватить 429. Если всё же схватили — auto-retry до 3 раз с задержкой из заголовка Retry-After / x-ratelimit-reset (или парс "try again in 6.7s" из текста). Если ждать > 30 сек — отказываемся, показываем ошибку. UI пишет: rate limit · повтор через 6.7s (попытка 2/4).',
    },
    {
      q: 'Безопасно ли запускать curl | sudo bash для установки?',
      a: 'Скрипт короткий, читай его перед запуском: github.com/kitay-sudo/wtf/blob/main/install.sh. Он только определяет архитектуру, скачивает бинарь из GitHub Releases и кладёт в /usr/local/bin/wtf. Никаких внешних серверов кроме github.com.',
    },
    {
      q: 'Можно ли работать офлайн?',
      a: 'Нет — AI-провайдеры все в облаке, и для tool-use агента кеш не работает (каждая сессия — новая последовательность команд). Self-hosted режим через ollama в roadmap.',
    },
    {
      q: 'А если я в Windows?',
      a: 'Работает через PowerShell (Windows PowerShell 5.1 и pwsh 7+) — команды агент запускает через cmd.exe / pwsh. На Linux/macOS — через /bin/sh. Также работает через Git Bash / WSL.',
    },
    {
      q: 'Можно прервать диагностику?',
      a: 'Да, Ctrl+C — graceful shutdown. Агент завершит текущий HTTP-запрос, сохранит память что успел собрать, и выйдет с кодом 130 (стандарт SIGINT).',
    },
    {
      q: 'Что если я хочу увидеть полный вывод команд?',
      a: 'По умолчанию вывод скрыт — каждая команда сворачивается в одну строку с временем, статусом, размером. Запусти с флагом -v или --verbose — увидишь полный вывод каждой команды. При exit≠0 в обычном режиме автоматически показываются последние 5 строк.',
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
            Перестань гуглить команды для починки.<br />
            <span className="bg-gradient-to-r from-amber-400 to-yellow-300 bg-clip-text text-transparent">
              Опиши проблему — wtf разрулит.
            </span>
          </h2>
          <p className="mt-5 text-zinc-400 max-w-xl mx-auto">
            60 секунд от установки до первой решённой проблемы. Один бинарь, твой ключ AI, никаких подписок.
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
    version: '0.2.0',
    date: '2026-05-01',
    title: 'Терминальный sysadmin-агент',
    highlights: [
      'Полностью новая концепция: пишешь wtf <вопрос словами> — агент сам разруливает',
      'Tool-use API всех трёх провайдеров (Claude / OpenAI / Gemini): надёжный парсинг вызовов',
      'Exec-classifier: 80+ безопасных read-only утилит для авто-запуска, blacklist destructive',
      'Память между сессиями: ~/.wtf/memory/store.json с TTL и AI-консолидацией раз в 20 сессий',
      'Auto-retry при 429: парс Retry-After / x-ratelimit-reset, до 3 попыток',
      'Quiet UI: каждая строка с [HH:MM:SS], спиннер в позиции иконки результата',
      'Throttle 800мс между AI-раундами + trim истории (4 последних tool_result)',
      'Graceful Ctrl+C: память сохраняется при прерывании',
      'Удалены: shell-хуки (wtf init, wtfc), кэш, --rerun, --explain — больше не нужны',
    ],
  },
  {
    version: '0.1.0',
    date: '2026-04-30',
    title: 'Первый публичный релиз',
    highlights: [
      'CLI на Go: один кросс-платформенный бинарник',
      'Три провайдера на выбор: Claude, OpenAI, Gemini',
      'Объяснение последнего вывода терминала через shell-хук',
      'Локальный кеш с TTL 30 дней, чистка секретов (13 regex-правил)',
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
              wtf пилится в свободное время, без подписок и платных тарифов.
              Если он сэкономил тебе пару часов и хочется отблагодарить — буду рад крипто-донату.
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

        <Reveal delay={0.1}>
          <div className="mt-10 grid grid-cols-1 md:grid-cols-2 gap-4">
            {wallets.map((w, i) => (
              <WalletCard key={w.label} {...w} delay={i * 0.05} />
            ))}
          </div>
        </Reveal>

        <Reveal delay={0.18}>
          <TimewebCard />
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

// TimewebCard — личная рекомендация хостинга. Пользователь приходит сюда
// потому что ему нужен VPS, а не потому что мы просим. Бейдж "ad" остаётся
// для честности (это партнёрская ссылка по требованиям FTC/ФЗ "О рекламе"),
// но сам месседж — про то что хостинг хороший.
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
              VPS, который мы используем сами
            </h3>
            <span className="inline-flex items-center px-1.5 py-0.5 rounded border border-zinc-700 bg-zinc-900/60 text-[10px] uppercase tracking-widest font-mono text-zinc-400">
              ad
            </span>
          </div>
          <p className="text-sm text-zinc-400 leading-relaxed">
            <a
              href="https://timeweb.cloud/?i=104289"
              target="_blank"
              rel="sponsored noopener"
              className="text-amber-300 hover:text-amber-200 font-medium underline-offset-4 hover:underline"
            >
              Timeweb Cloud
            </a>{' '}
            — российский хостинг, на котором живут наши боевые сервера: быстрая
            панель, NVMe-диски, развёртывание VPS за минуту, оплата картой и
            крипто-кошельком. Ровно то, что нужно когда ты деплоишь свои сервисы
            и хочешь чтобы wtf работал по SSH без задержек.
          </p>
          <p className="mt-3 text-sm text-zinc-400 leading-relaxed">
            Берёшь сервер — могу{' '}
            <span className="text-zinc-200">помочь с первичной настройкой</span>:
            напиши в Telegram{' '}
            <a
              href="https://t.me/kitay9"
              target="_blank"
              rel="noreferrer"
              className="text-amber-300 hover:text-amber-200 font-mono"
            >
              @kitay9
            </a>{' '}
            — подскажу с конфигом, файрволом, systemd, деплоем своих сервисов.
          </p>
          <a
            href="https://timeweb.cloud/?i=104289"
            target="_blank"
            rel="sponsored noopener"
            className="mt-5 inline-flex items-center gap-1.5 text-sm font-medium px-4 py-2 rounded-lg bg-amber-400 hover:bg-amber-300 text-zinc-950 transition-colors"
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
