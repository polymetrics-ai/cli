import Link from 'next/link';
import { CornerBox } from '@/components/ui/corner-box';

const REPO = 'https://github.com/polymetrics-ai/cli';

/* ── Link columns ─────────────────────────────────────────────────────── */
type Col = { title: string; links: [label: string, href: string][] };

const COLUMNS: Col[] = [
  {
    title: 'Product',
    links: [
      ['Extract (ETL)', '/docs/connectors'],
      ['Query (SQL)', '/docs/query'],
      ['Reverse-ETL', '/docs/reverse-etl'],
      ['Agent Mode', '/docs/agent'],
      ['Local Vault', '/docs/vault'],
    ],
  },
  {
    title: 'Developers',
    links: [
      ['Documentation', '/docs'],
      ['Quickstart', '/docs/quickstart'],
      ['Connectors', '/docs/connectors'],
      ['Bidirectional', '/docs/bidirectional'],
      ['CLI Reference', '/docs'],
    ],
  },
  {
    title: 'Resources',
    links: [
      ['Changelog', '/changelog'],
      ['Pattern registry', '/patterns'],
      ['GitHub', REPO],
      ['Issues', `${REPO}/issues`],
      ['Discussions', `${REPO}/discussions`],
    ],
  },
  {
    title: 'Community',
    links: [
      ['Public Source', REPO],
      ['Elastic License 2.0', `${REPO}/blob/main/LICENSE`],
      ['Star on GitHub', REPO],
      ['Report a bug', `${REPO}/issues/new`],
    ],
  },
];

function FooterLink({ label, href }: { label: string; href: string }) {
  const cls =
    'text-[13px] text-text-tertiary hover:text-text-primary transition-colors';
  return href.startsWith('http') ? (
    <a href={href} target="_blank" rel="noreferrer" className={cls}>
      {label}
    </a>
  ) : (
    <Link href={href} className={cls}>
      {label}
    </Link>
  );
}

/* ── Social icons (inline brand SVGs) ─────────────────────────────────── */
const SOCIALS: { label: string; href: string; path: string }[] = [
  {
    label: 'GitHub',
    href: REPO,
    path: 'M12 .5C5.37.5 0 5.78 0 12.29c0 5.2 3.44 9.6 8.21 11.16.6.11.82-.25.82-.56 0-.28-.01-1.02-.02-2-3.34.71-4.04-1.58-4.04-1.58-.55-1.37-1.34-1.74-1.34-1.74-1.09-.73.08-.72.08-.72 1.2.08 1.84 1.21 1.84 1.21 1.07 1.79 2.81 1.27 3.5.97.11-.76.42-1.27.76-1.56-2.67-.3-5.47-1.31-5.47-5.83 0-1.29.47-2.34 1.24-3.17-.12-.3-.54-1.52.12-3.16 0 0 1.01-.32 3.3 1.21a11.6 11.6 0 0 1 3-.4c1.02 0 2.05.13 3 .4 2.29-1.53 3.3-1.21 3.3-1.21.66 1.64.24 2.86.12 3.16.77.83 1.24 1.88 1.24 3.17 0 4.53-2.81 5.53-5.49 5.82.43.37.81 1.1.81 2.22 0 1.6-.01 2.9-.01 3.29 0 .31.22.68.82.56A12.01 12.01 0 0 0 24 12.29C24 5.78 18.63.5 12 .5Z',
  },
  {
    label: 'X',
    href: '#',
    path: 'M18.244 2.25h3.308l-7.227 8.26 8.502 11.24h-6.66l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231 5.45-6.231Zm-1.161 17.52h1.833L7.084 4.126H5.117L17.083 19.77Z',
  },
  {
    label: 'LinkedIn',
    href: '#',
    path: 'M20.45 20.45h-3.56v-5.57c0-1.33-.03-3.04-1.85-3.04-1.85 0-2.13 1.45-2.13 2.94v5.67H9.35V9h3.42v1.56h.05c.48-.9 1.64-1.85 3.37-1.85 3.6 0 4.27 2.37 4.27 5.46v6.28ZM5.34 7.43a2.06 2.06 0 1 1 0-4.13 2.06 2.06 0 0 1 0 4.13ZM7.12 20.45H3.55V9h3.57v11.45ZM22.22 0H1.77C.79 0 0 .77 0 1.73v20.54C0 23.23.79 24 1.77 24h20.45c.98 0 1.78-.77 1.78-1.73V1.73C24 .77 23.2 0 22.22 0Z',
  },
  {
    label: 'Discord',
    href: '#',
    path: 'M20.317 4.369a19.79 19.79 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.74 19.74 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028c.462-.63.874-1.295 1.226-1.994a.076.076 0 0 0-.041-.106 13.1 13.1 0 0 1-1.872-.892.077.077 0 0 1-.008-.128c.126-.094.252-.192.372-.291a.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.061 0a.074.074 0 0 1 .078.009c.12.099.246.198.373.292a.077.077 0 0 1-.006.127c-.598.349-1.22.645-1.873.892a.076.076 0 0 0-.04.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.84 19.84 0 0 0 6.002-3.03.077.077 0 0 0 .032-.056c.5-5.177-.838-9.674-3.549-13.66a.06.06 0 0 0-.031-.028ZM8.02 15.331c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418Zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418Z',
  },
  {
    label: 'YouTube',
    href: '#',
    path: 'M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814ZM9.545 15.568V8.432L15.818 12l-6.273 3.568Z',
  },
];

function SocialIcon({ label, href, path }: { label: string; href: string; path: string }) {
  return (
    <a
      href={href}
      target={href.startsWith('http') ? '_blank' : undefined}
      rel={href.startsWith('http') ? 'noreferrer' : undefined}
      aria-label={label}
      className="text-text-tertiary hover:text-text-primary transition-colors"
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
        <path d={path} />
      </svg>
    </a>
  );
}

/* ── Footer — sits inside the center grid, bracketed like the sections ── */
export function SiteFooter() {
  return (
    <footer className="mx-auto w-full px-4 sm:px-8 md:px-0 md:max-w-[680px] xl:max-w-[840px] pt-[120px]">
      <CornerBox>
        {/* Social row */}
        <div className="flex items-center gap-5 px-6 py-5">
          {SOCIALS.map((s) => (
            <SocialIcon key={s.label} {...s} />
          ))}
        </div>

        {/* Link columns */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-x-6 gap-y-8 border-t border-line-structure px-6 py-8">
          {COLUMNS.map((col) => (
            <div key={col.title} className="flex flex-col gap-3">
              <span className="text-[12px] font-medium text-text-disabled">{col.title}</span>
              <ul className="flex flex-col gap-2.5">
                {col.links.map(([label, href]) => (
                  <li key={label}>
                    <FooterLink label={label} href={href} />
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Legal links */}
        <div className="flex flex-wrap items-center gap-5 border-t border-line-structure px-6 py-4 text-[12px]">
          <FooterLink label="Elastic License 2.0" href={`${REPO}/blob/main/LICENSE`} />
          <FooterLink label="Privacy" href="#" />
          <FooterLink label="Security" href={`${REPO}/security`} />
        </div>

        {/* Copyright + trademark footnote */}
        <div className="flex flex-col gap-3 border-t border-line-structure px-6 py-4">
          <div className="flex flex-wrap items-center justify-between gap-2 text-[12px] text-text-tertiary">
            <span>© 2026 Polymetrics AI, public source project.</span>
            <span className="flex items-center gap-2">
              <span className="flex items-center justify-center h-[18px] min-w-[18px] px-1 bg-emerald-800 select-none">
                <span className="font-mono font-bold text-[10px] leading-none text-white tracking-tight">PM</span>
                <span aria-hidden className="font-mono font-bold text-[10px] leading-none text-white cursor-blink">_</span>
              </span>
              Built in pure Go.
            </span>
          </div>
          <p className="text-[11px] leading-relaxed text-text-disabled max-w-[90ch]">
            All product names, logos, trademarks, and brand names are the property of their
            respective owners. Connector names are shown for identification and
            interoperability only; their listing does not imply any affiliation with or
            endorsement by their respective owners.
          </p>
        </div>
      </CornerBox>
    </footer>
  );
}
