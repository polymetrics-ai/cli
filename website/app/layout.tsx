import { RootProvider } from 'fumadocs-ui/provider/next';
import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { Geist, Geist_Mono, Instrument_Serif, Chakra_Petch } from 'next/font/google';
import { TooltipProvider } from '@/components/ui/tooltip';
import { CONNECTOR_CATALOG_COUNT } from '@/lib/connectors.generated';
import './globals.css';

const geistSans = Geist({
  subsets: ['latin'],
  variable: '--font-geist-sans',
  display: 'swap',
});

// Thin, squared technical face — used for the "command line interface" mark.
const chakraPetch = Chakra_Petch({
  subsets: ['latin'],
  weight: ['300'],
  variable: '--font-chakra',
  display: 'swap',
});

const geistMono = Geist_Mono({
  subsets: ['latin'],
  variable: '--font-geist-mono',
  display: 'swap',
});

const instrumentSerif = Instrument_Serif({
  subsets: ['latin'],
  weight: ['400'],
  style: ['normal', 'italic'],
  variable: '--font-analog',
  display: 'swap',
});

export const metadata: Metadata = {
  title: {
    default: 'Polymetrics CLI: One CLI to rule them all',
    template: '%s | pm · Polymetrics',
  },
  description:
    `pm is a local-first, single-binary data engine. Browse ${CONNECTOR_CATALOG_COUNT} connectors, query with embedded DuckDB SQL, and write results back. No Docker. Agent-native.`,
  metadataBase: new URL('https://cli.polymetrics.ai'),
  openGraph: {
    siteName: 'Polymetrics CLI',
    title: 'Polymetrics CLI: One CLI to rule them all',
    description:
      `Local-first ETL, DuckDB SQL, reverse ETL, and AI-agent-safe automation across ${CONNECTOR_CATALOG_COUNT} connectors.`,
    url: 'https://cli.polymetrics.ai',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Polymetrics CLI: One CLI to rule them all',
    description:
      'A local-first data CLI for ETL, DuckDB SQL, reverse ETL, and agent-native automation.',
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={`light ${geistSans.variable} ${geistMono.variable} ${instrumentSerif.variable} ${chakraPetch.variable}`}
    >
      <body className="font-sans flex min-h-screen flex-col antialiased">
        <RootProvider search={{ enabled: false }} theme={{ enabled: false }}>
          <TooltipProvider>{children}</TooltipProvider>
        </RootProvider>
      </body>
    </html>
  );
}
