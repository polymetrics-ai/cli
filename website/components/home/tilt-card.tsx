'use client';

import { useRef, useCallback, type MouseEvent } from 'react';
import { cn } from '@/lib/utils';

interface TiltCardProps {
  children: React.ReactNode;
  className?: string;
  intensity?: number;
}

export function TiltCard({ children, className, intensity = 8 }: TiltCardProps) {
  const ref = useRef<HTMLDivElement>(null);
  const raf = useRef<number | undefined>(undefined);
  const isInside = useRef(false);

  const move = useCallback(
    (e: MouseEvent) => {
      const el = ref.current;
      if (!el) return;
      isInside.current = true;
      if (raf.current) cancelAnimationFrame(raf.current);
      raf.current = requestAnimationFrame(() => {
        const { left, top, width, height } = el.getBoundingClientRect();
        const x = ((e.clientX - left) / width - 0.5) * 2;
        const y = ((e.clientY - top) / height - 0.5) * 2;
        el.style.transform = `perspective(900px) rotateY(${x * intensity}deg) rotateX(${-y * intensity}deg) translateZ(6px)`;
        el.style.boxShadow = `0 ${12 + Math.abs(y) * 20}px ${28 + Math.abs(y) * 40}px -8px rgba(64,64,57,${0.1 + Math.abs(y) * 0.08})`;
      });
    },
    [intensity],
  );

  const leave = useCallback(() => {
    const el = ref.current;
    if (!el) return;
    isInside.current = false;
    el.style.transition = 'transform 0.7s cubic-bezier(0.23,1,0.32,1), box-shadow 0.5s ease';
    el.style.transform = '';
    el.style.boxShadow = '';
    setTimeout(() => {
      if (!isInside.current && el) el.style.transition = '';
    }, 700);
  }, []);

  return (
    <div
      ref={ref}
      className={cn(className)}
      onMouseMove={move}
      onMouseLeave={leave}
      style={{ willChange: 'transform', transition: 'transform 0.1s ease-out, box-shadow 0.1s ease-out', transformStyle: 'preserve-3d' }}
    >
      {children}
    </div>
  );
}
