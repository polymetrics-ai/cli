import { HomeLayout } from 'fumadocs-ui/layouts/home';
import type { ReactNode } from 'react';
import { baseOptions } from '@/lib/layout.shared';
import { SiteNavbar } from '@/components/home/navbar';

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <HomeLayout {...baseOptions} slots={{ header: SiteNavbar }}>
      {children}
    </HomeLayout>
  );
}
