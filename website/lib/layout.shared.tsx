import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';

export const baseOptions: BaseLayoutProps = {
  nav: {
    title: (
      <span className="flex items-center justify-center h-[24px] min-w-[24px] px-1 bg-emerald-800 select-none">
        <span className="font-mono font-bold text-[12px] leading-none text-white tracking-tight">PM</span>
        <span aria-hidden className="font-mono font-bold text-[12px] leading-none text-white cursor-blink">_</span>
      </span>
    ),
  },
  links: [
    {
      text: 'Documentation',
      url: '/docs',
      active: 'nested-url',
    },
  ],
  githubUrl: 'https://github.com/polymetrics-ai/cli',
};
