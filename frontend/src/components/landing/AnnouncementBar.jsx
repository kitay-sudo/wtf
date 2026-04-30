import { useEffect, useState } from 'react';
import { X, Server, ArrowRight } from 'lucide-react';

// Тонкая полоска с партнёрской ссылкой Timeweb. Прячется в localStorage когда
// пользователь закрыл — показываем заново только если меняем ключ.
//
// Disclosure: это явно партнёрская ссылка. Прячем сам факт партнёрки только тогда,
// когда юзер её закрыл (а не саму суть). Слово "ad" / "sponsor" видно сразу.
const STORAGE_KEY = 'wtf-announcement-timeweb-1';
const REF_URL = 'https://timeweb.cloud/?i=104289';

export default function AnnouncementBar() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    if (window.localStorage.getItem(STORAGE_KEY) !== 'closed') {
      setVisible(true);
    }
  }, []);

  const close = () => {
    setVisible(false);
    try {
      window.localStorage.setItem(STORAGE_KEY, 'closed');
    } catch {
      /* private mode etc — ничего страшного */
    }
  };

  if (!visible) return null;

  return (
    <div className="relative bg-amber-400/10 border-b border-amber-400/20 text-zinc-200">
      <div className="max-w-6xl mx-auto px-5 py-2 flex items-center justify-center gap-3 text-xs sm:text-sm">
        <span className="hidden sm:inline-flex items-center justify-center w-5 h-5 rounded bg-amber-400/15 text-amber-300 shrink-0">
          <Server size={12} />
        </span>
        <span className="text-zinc-300">
          <span className="text-zinc-500 font-mono mr-2 hidden sm:inline">[ad]</span>
          Поднимаешь сервер для <code className="font-mono text-amber-300">wtf</code>?{' '}
          <a
            href={REF_URL}
            target="_blank"
            rel="sponsored noopener"
            className="text-amber-300 hover:text-amber-200 font-medium underline-offset-4 hover:underline inline-flex items-center gap-1"
          >
            Timeweb Cloud
            <ArrowRight size={12} />
          </a>{' '}
          <span className="text-zinc-500">— по нашей ссылке поможем с настройкой.</span>
        </span>
        <button
          onClick={close}
          aria-label="Закрыть"
          className="ml-2 shrink-0 p-1 rounded hover:bg-amber-400/10 text-zinc-500 hover:text-zinc-200 transition-colors"
        >
          <X size={14} />
        </button>
      </div>
    </div>
  );
}
