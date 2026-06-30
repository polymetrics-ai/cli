'use client';

import { cn } from '@/lib/utils';
import React, { forwardRef, useEffect, useRef, useState } from 'react';
import { MathBackdrop } from '@/components/home/math-backdrop';

type Props = React.HTMLAttributes<HTMLElement> & {
  /** Pattern id from the math registry to paint faintly behind this section. */
  pattern?: string;
};

export const HomeSection = forwardRef<HTMLElement, Props>(
  ({ className, children, pattern, ...props }, ref) => {
    const innerRef = useRef<HTMLElement>(null);
    // Start visible so SSR / full-page screenshots show content.
    const [mounted, setMounted] = useState(false);
    const [animated, setAnimated] = useState(false);

    useEffect(() => {
      setMounted(true);
      const el = innerRef.current;
      if (!el) return;

      const rect = el.getBoundingClientRect();
      const inView = rect.top < window.innerHeight && rect.bottom > 0;
      if (inView) return;

      setAnimated(true);
      const observer = new IntersectionObserver(
        ([entry]) => {
          if (entry.isIntersecting) {
            setAnimated(false);
            observer.disconnect();
          }
        },
        { threshold: 0, rootMargin: '0px 0px -40px 0px' },
      );
      observer.observe(el);
      return () => observer.disconnect();
    }, []);

    return (
      <section
        ref={(node) => {
          (innerRef as React.MutableRefObject<HTMLElement | null>).current = node;
          if (typeof ref === 'function') ref(node);
          else if (ref) (ref as React.MutableRefObject<HTMLElement | null>).current = node;
        }}
        className={cn(
          'relative w-full',
          mounted && animated
            ? 'opacity-0 translate-y-4 transition-[opacity,transform] duration-700 ease-[cubic-bezier(0.25,0.1,0.25,1)]'
            : mounted
              ? 'opacity-100 translate-y-0 transition-[opacity,transform] duration-700 ease-[cubic-bezier(0.25,0.1,0.25,1)]'
              : 'opacity-100',
          className,
        )}
        {...props}
      >
        {pattern && <MathBackdrop id={pattern} fade="edges" caption />}
        {/* Content column — same constraint as before, lifted above the backdrop */}
        <div className="relative z-[1] mx-auto w-full px-4 sm:px-8 md:px-0 md:max-w-[680px] xl:max-w-[840px]">
          {children}
        </div>
      </section>
    );
  },
);
HomeSection.displayName = 'HomeSection';
