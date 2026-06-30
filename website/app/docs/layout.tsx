import type { ReactNode } from 'react';
import { SiteNavbar } from '@/components/home/navbar';
import { DocsSidebar } from '@/components/docs/docs-sidebar';
import { DocsTocAside } from '@/components/docs/docs-toc';

export default function DocsLayoutWrapper({ children }: { children: ReactNode }) {
  return (
    <>
      <SiteNavbar />
      <div className="flex mx-auto w-full max-w-[95rem] overflow-clip">
        <DocsSidebar />
        <main
          id="nd-docs-layout"
          className="flex-1 min-w-0 pattern-bg flex flex-col"
        >
          {children}
        </main>
        <DocsTocAside />
      </div>
    </>
  );
}
