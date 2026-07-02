import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';
import { PmLogoMark } from '@/components/brand/pm-logo-mark';

export const baseOptions: BaseLayoutProps = {
  nav: {
    title: <PmLogoMark className="h-6 w-6 shrink-0 select-none" />,
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
