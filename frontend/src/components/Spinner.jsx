import { useEffect, useState } from 'react';

// Braille-frames like real CLI spinners (cli-spinners "dots" preset).
// Меняются по таймеру, не CSS-rotate — выглядит как настоящий терминал.
const FRAMES = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];

export default function Spinner({ className = '', interval = 80 }) {
  const [i, setI] = useState(0);
  useEffect(() => {
    const id = setInterval(() => setI((x) => (x + 1) % FRAMES.length), interval);
    return () => clearInterval(id);
  }, [interval]);
  return (
    <span className={`inline-block font-mono ${className}`} aria-hidden>
      {FRAMES[i]}
    </span>
  );
}
