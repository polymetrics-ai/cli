import { NextResponse } from 'next/server';
import { NextRequest } from 'next/server';
import { connectorBySlug, CONNECTOR_CATALOG } from '@/lib/connectors.catalog.generated';

export function generateStaticParams() {
  return CONNECTOR_CATALOG.map((c) => ({ slug: c.slug }));
}

export async function GET(
  _req: NextRequest,
  { params }: { params: Promise<{ slug: string }> },
) {
  const { slug } = await params;
  const connector = connectorBySlug(slug);

  if (!connector) {
    return new NextResponse(JSON.stringify({ error: 'Connector not found' }), {
      status: 404,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  return NextResponse.json(connector);
}
