// Лого утилиты: "!?" в монопространственном шрифте на тёмном фоне
// с жёлтой обводкой. Это же — favicon.
export default function WtfMark({ size = 28, className = '', strokeOpacity = 0.55 }) {
  const fontSize = Math.round(size * 0.5);
  return (
    <span
      className={`inline-flex items-center justify-center select-none ${className}`}
      style={{ width: size, height: size }}
      aria-label="wtf"
    >
      <svg width={size} height={size} viewBox="0 0 32 32" fill="none" aria-hidden>
        <rect width="32" height="32" rx="7" fill="#1c1917" />
        <rect
          x="0.5"
          y="0.5"
          width="31"
          height="31"
          rx="6.5"
          stroke="#fbbf24"
          strokeOpacity={strokeOpacity}
        />
        <text
          x="16"
          y="22"
          fontFamily="ui-monospace, 'JetBrains Mono', SFMono-Regular, Menlo, monospace"
          fontWeight="800"
          fontSize="15"
          letterSpacing="-0.5"
          textAnchor="middle"
          fill="#fbbf24"
        >
          !?
        </text>
      </svg>
    </span>
  );
}
