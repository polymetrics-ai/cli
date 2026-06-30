import type { Metadata } from 'next';
import { ConnectorBrowser } from '@/components/docs/connector-browser';
import { CONNECTOR_CATALOG_COUNT } from '@/lib/connectors.catalog.generated';

export const metadata: Metadata = {
  title: 'Connectors',
  description: `Browse and search all ${CONNECTOR_CATALOG_COUNT} pm connectors — sources and destinations across APIs, databases, files, and more.`,
};

export default function ConnectorsIndexPage() {
  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-10">
      <ConnectorBrowser />
    </div>
  );
}
