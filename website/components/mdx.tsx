import { isValidElement, type ReactNode } from 'react';
import defaultComponents from 'fumadocs-ui/mdx';
import { Card, Cards } from 'fumadocs-ui/components/card';
import { Steps, Step } from 'fumadocs-ui/components/steps';
import { Tabs, Tab } from 'fumadocs-ui/components/tabs';
import { Accordion, Accordions } from 'fumadocs-ui/components/accordion';
import { DocsCallout } from '@/components/docs/docs-callout';
import { DocsCodeBlock } from '@/components/docs/docs-code-block';
import { ExtractQueryActDiagram } from '@/components/docs/extract-query-act-diagram';

function textFromNode(node: ReactNode): string {
  if (typeof node === 'string' || typeof node === 'number') return String(node);
  if (Array.isArray(node)) return node.map(textFromNode).join('');
  if (isValidElement<{ children?: ReactNode }>(node)) return textFromNode(node.props.children);
  return '';
}

function languageFromNode(node: ReactNode): string | undefined {
  if (!isValidElement<{ className?: string; 'data-language'?: string }>(node)) return undefined;
  const dataLanguage = node.props['data-language'];
  if (typeof dataLanguage === 'string') return dataLanguage;
  const match = node.props.className?.match(/language-([a-z0-9_-]+)/i);
  return match?.[1];
}

function Pre({ children }: { children?: ReactNode }) {
  const code = textFromNode(children);
  const language = languageFromNode(children);

  if (!code.trim()) {
    return <pre>{children}</pre>;
  }

  return <DocsCodeBlock code={code} language={language} />;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function getMDXComponents(extra?: Record<string, any>): Record<string, any> {
  return {
    ...defaultComponents,
    pre: Pre,
    Card,
    Cards,
    Callout: DocsCallout,
    Steps,
    Step,
    Tabs,
    Tab,
    Accordion,
    Accordions,
    ExtractQueryActDiagram,
    ...extra,
  };
}
