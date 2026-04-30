// Маленький "значок" для лого — вместо изображения используем эмодзи 🤬,
// центрированный в квадратике. Это бренд-знак утилиты wtf.
export default function RageMark({ size = 28, className = '' }) {
  return (
    <span
      className={`inline-flex items-center justify-center leading-none ${className}`}
      style={{ width: size, height: size, fontSize: Math.round(size * 0.75) }}
      aria-label="wtf"
    >
      🤬
    </span>
  );
}
