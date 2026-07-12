import {
  Cable,
  Database,
  ExternalLink,
  FileJson,
  GitBranch as Github,
  HardDrive,
  Lock,
  Newspaper,
  PackageCheck,
  Scale,
  Sparkles,
  Star,
} from 'lucide-react';
import { SidebarLink } from '@/components/home/sidebar-link';
import { DOCUMENTATION_NAV } from '@/components/docs/doc-nav';
import { BLOG_POSTS, blogUrl } from '@/lib/blog';
import { CONNECTOR_CATALOG_COUNT } from '@/lib/connectors.generated';
import {
  Sidebar,
  SidebarAccent,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupHeader,
  SidebarGroupLabel,
  SidebarInner,
  SidebarMenu,
  SidebarStat,
} from '@/components/ui/sidebar';

const stats = [
  {
    label: 'GitHub Stars',
    value: 'Star us',
    href: 'https://github.com/polymetrics-ai/cli',
    tooltip: 'Leave a Star',
    icon: Star,
  },
  {
    label: 'Connectors',
    value: String(CONNECTOR_CATALOG_COUNT),
    href: '/docs/connectors',
    tooltip: 'Browse connectors',
    icon: Cable,
  },
  {
    label: 'License',
    value: 'Elastic-2.0',
    href: 'https://github.com/polymetrics-ai/cli/blob/main/LICENSE',
    tooltip: 'Elastic License 2.0',
    icon: Scale,
  },
  {
    label: 'Binary size',
    value: '~18MB',
    href: '/docs/installation',
    tooltip: 'Single binary install',
    icon: HardDrive,
  },
];

const changelog = [
  { version: 'release', label: 'Automated binary releases', icon: PackageCheck },
  { version: 'v0.1.0', label: `${CONNECTOR_CATALOG_COUNT} connector catalog pages`, icon: Cable },
  { version: 'v0.1.0', label: 'Embedded DuckDB SQL engine', icon: Database },
  { version: 'v0.1.0', label: 'Local AES-GCM vault', icon: Lock },
];

const latestBlogPosts = BLOG_POSTS.slice(0, 3);

export function HomeSidebar({
  className = 'home-sidebar-panel',
}: {
  className?: string;
}) {
  return (
    <Sidebar className={className}>
      <SidebarInner>
        <SidebarContent className="space-y-2 pt-3">
          <SidebarGroup>
            <SidebarGroupHeader>
              <SidebarGroupLabel>Project stats</SidebarGroupLabel>
              <Sparkles className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
            </SidebarGroupHeader>
            <SidebarGroupContent>
              <SidebarMenu>
                {stats.map(({ label, value, href, tooltip, icon: Icon }) => (
                  <SidebarLink key={label} href={href} tooltip={tooltip} className="p-0">
                    <SidebarStat
                      label={
                        <span className="flex min-w-0 items-center gap-1.5">
                          <Icon className="h-3.5 w-3.5 shrink-0 text-text-disabled transition-colors group-hover:text-line-cta" aria-hidden="true" />
                          <span className="truncate">{label}</span>
                        </span>
                      }
                      value={value}
                    />
                  </SidebarLink>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>

          <SidebarGroup>
            <SidebarGroupHeader>
              <SidebarGroupLabel>Documentation</SidebarGroupLabel>
              <FileJson className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
            </SidebarGroupHeader>
            <SidebarGroupContent>
              <SidebarMenu>
                {DOCUMENTATION_NAV.map(({ label, href, icon: Icon }) => (
                  <SidebarLink key={label} href={href}>
                    <span className="flex min-w-0 items-center gap-2">
                      <span className="flex h-5 w-5 shrink-0 items-center justify-center border border-transparent transition-colors group-hover:border-line-structure group-hover:bg-surface-1">
                        <Icon className="h-3.5 w-3.5 text-text-disabled transition-colors group-hover:text-line-cta" aria-hidden="true" />
                      </span>
                      <span className="truncate text-[13px] leading-snug text-text-tertiary transition-colors group-hover:text-text-secondary">
                        {label}
                      </span>
                    </span>
                  </SidebarLink>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>

          <SidebarGroup>
            <SidebarGroupHeader>
              <SidebarGroupLabel>Blog</SidebarGroupLabel>
              <Newspaper className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
            </SidebarGroupHeader>
            <SidebarGroupContent>
              <SidebarMenu>
                {latestBlogPosts.map((post) => (
                  <SidebarLink key={post.slug} href={blogUrl(post.slug)}>
                    <span className="flex min-w-0 items-start gap-2">
                      <span className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center border border-transparent transition-colors group-hover:border-line-structure group-hover:bg-surface-1">
                        <Newspaper className="h-3.5 w-3.5 text-text-disabled transition-colors group-hover:text-line-cta" aria-hidden="true" />
                      </span>
                      <span className="min-w-0">
                        <span className="block truncate text-[13px] leading-snug text-text-tertiary transition-colors group-hover:text-text-secondary">
                          {post.title}
                        </span>
                        <span className="mt-0.5 block font-mono text-[10px] uppercase tracking-wider text-text-disabled">
                          {post.readingTime}
                        </span>
                      </span>
                    </span>
                  </SidebarLink>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>

          <SidebarGroup>
            <SidebarGroupHeader>
              <SidebarGroupLabel>What&rsquo;s new</SidebarGroupLabel>
              <a
                href="https://github.com/polymetrics-ai/cli/releases"
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-1 font-mono text-[10px] uppercase tracking-wider text-text-tertiary transition-colors hover:text-text-secondary"
              >
                View all
                <ExternalLink className="h-3 w-3" aria-hidden="true" />
              </a>
            </SidebarGroupHeader>
            <SidebarGroupContent className="space-y-2.5">
              {changelog.map(({ version, label, icon: Icon }) => (
                <div key={label} className="grid grid-cols-[1.1rem_minmax(0,1fr)] gap-2">
                  <span className="mt-0.5 flex h-4 w-4 items-center justify-center border border-line-structure bg-surface-1">
                    <Icon className="h-2.5 w-2.5 text-line-cta" aria-hidden="true" />
                  </span>
                  <span className="min-w-0">
                    <span className="block text-[12px] leading-snug text-text-secondary">
                      {label}
                    </span>
                    <span className="mt-0.5 block font-mono text-[10px] text-text-disabled">
                      {version} · today
                    </span>
                  </span>
                </div>
              ))}
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter>
          <a
            href="https://github.com/polymetrics-ai/cli"
            target="_blank"
            rel="noreferrer"
            className="link-box group relative grid min-w-0 grid-cols-[2rem_minmax(0,1fr)] items-center gap-2 px-4 py-3 text-[12px] text-text-tertiary transition-colors hover:bg-surface-bg hover:text-text-secondary"
          >
            <span aria-hidden className="corner-box-hover-child" />
            <span className="flex h-8 w-8 items-center justify-center bg-emerald-800 font-mono text-[13px] font-bold leading-none text-white">
              PM<span aria-hidden className="cursor-blink">_</span>
            </span>
            <span className="min-w-0">
              <span className="block font-square text-[11px] font-semibold uppercase tracking-wider text-text-secondary">
                pm CLI
              </span>
              <span className="mt-0.5 flex min-w-0 items-center gap-1.5 font-mono text-[10px] text-text-disabled">
                <Github className="h-3 w-3 shrink-0" aria-hidden="true" />
                <span className="min-w-0 truncate">polymetrics-ai/cli</span>
              </span>
            </span>
          </a>
          <SidebarAccent />
        </SidebarFooter>
      </SidebarInner>
    </Sidebar>
  );
}
