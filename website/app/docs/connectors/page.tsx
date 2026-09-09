import type { Metadata } from 'next';
import Link from 'next/link';
import { ConnectorBrowser } from '@/components/docs/connector-browser';
import { CONNECTOR_CATALOG_COUNT } from '@/lib/connectors.catalog.generated';

export const metadata: Metadata = {
  title: 'Connectors',
  description: `Browse and search all ${CONNECTOR_CATALOG_COUNT} pm connector bundles across APIs, databases, files, and more.`,
};

export default function ConnectorsIndexPage() {
  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-10">
      <section aria-labelledby="source-inspection" className="mb-8 border border-line-structure bg-surface-bg p-5">
        <h2 id="source-inspection" className="mb-2 font-medium text-text-primary">Inspect retained source operations</h2>
        <p className="mb-3 max-w-3xl text-sm leading-relaxed text-text-tertiary">
          Use source inspection to find retained operations, their seven lane observations,
          and source-backed reasons. A source mapping can remain unproven while an existing
          command is supported. Inspection and preflight do not execute provider requests.
        </p>
        <pre className="mb-3 overflow-x-auto border border-line-structure p-3 text-xs text-text-primary"><code>pm connectors inspect asana --sources --json</code></pre>
        <Link href="/docs/cli-reference#retained-source-operations" className="text-sm underline underline-offset-4 text-text-primary">
          Source selection flags, typed outcomes, and examples
        </Link>
      </section>
      <ConnectorBrowser />
    </div>
  );
}
