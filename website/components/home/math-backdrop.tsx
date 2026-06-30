import type { CSSProperties } from 'react';
import { cn } from '@/lib/utils';
import { patternById } from '@/lib/patterns.generated';

type Fade = 'edges' | 'radial' | 'top' | 'none';

// Horizontal framing: clear the centre (reading area), show toward the gutters.
const FRAME = 'radial-gradient(ellipse 56% 88% at 50% 50%, transparent 22%, #000 76%)';

// Vertical fade so each section's pattern eases in at its top and out at its
// bottom. Adjacent sections therefore cross-fade through the boundary instead
// of meeting at a hard seam — the transition reads as smooth while scrolling.
const VERTICAL: Record<Fade, string | undefined> = {
  edges: 'linear-gradient(to bottom, transparent 0%, #000 24%, #000 76%, transparent 100%)',
  radial: 'linear-gradient(to bottom, transparent 0%, #000 12%, #000 88%, transparent 100%)',
  top: 'linear-gradient(to bottom, transparent 0%, #000 22%, #000 100%)',
  none: undefined,
};

const maskStyle = (img?: string): CSSProperties =>
  img
    ? {
        WebkitMaskImage: img,
        maskImage: img,
        WebkitMaskRepeat: 'no-repeat',
        maskRepeat: 'no-repeat',
        WebkitMaskSize: '100% 100%',
        maskSize: '100% 100%',
      }
    : {};

/**
 * A faint generative-math pattern from the registry, painted behind a section.
 * Two nested masks compose:
 *   outer = vertical fade  (smooth top/bottom -> cross-fades between sections)
 *   inner = horizontal frame (keeps the reading column calm)
 * Content sits on opaque cards above this.
 */
export function MathBackdrop({
  id,
  fade = 'edges',
  opacity = 0.85,
  caption = false,
  className,
}: {
  id: string;
  fade?: Fade;
  opacity?: number;
  caption?: boolean;
  className?: string;
}) {
  const p = patternById(id);
  const frame = fade === 'none' ? undefined : FRAME;
  return (
    <>
      <div
        aria-hidden
        className={cn('pointer-events-none absolute inset-0 z-0', className)}
        style={maskStyle(VERTICAL[fade])}
      >
        <div
          className="absolute inset-0 bg-center bg-no-repeat bg-cover [transition:opacity_300ms_ease]"
          style={{ backgroundImage: `url(${p.file})`, opacity, ...maskStyle(frame) }}
        />
      </div>
      {caption && (
        <span
          aria-hidden
          className="pointer-events-none absolute bottom-2 right-3 z-0 hidden font-mono text-[10px] text-text-disabled xl:block"
          title={p.name}
        >
          {p.formula}
        </span>
      )}
    </>
  );
}
