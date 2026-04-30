export default function SectionDivider({ symbol = '·', label }) {
  return (
    <div className="flex items-center justify-center gap-3 mb-4">
      <span className="h-px w-10 bg-gradient-to-r from-transparent to-zinc-700" />
      <span className="text-amber-400/80 text-base font-mono font-semibold">
        {symbol}
      </span>
      <span className="text-[11px] tracking-[0.25em] uppercase text-zinc-500 font-medium">
        {label}
      </span>
      <span className="h-px w-10 bg-gradient-to-l from-transparent to-zinc-700" />
    </div>
  );
}
