import { ArrowDown, ArrowLeft, ArrowRight, Bot, Cable, Database, GitBranch, SearchCode } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { Badge } from '@/components/ui/badge';

function DiagramNode({
  label,
  detail,
  icon: Icon,
}: {
  label: string;
  detail: string;
  icon: LucideIcon;
}) {
  return (
    <div className="border border-line-structure bg-surface-bg p-3">
      <div className="flex items-center gap-2">
        <span className="flex size-8 shrink-0 items-center justify-center border border-line-structure bg-surface-1 text-line-cta">
          <Icon className="size-4" aria-hidden="true" />
        </span>
        <span className="font-square text-[13px] font-semibold uppercase tracking-wider text-text-primary">
          {label}
        </span>
      </div>
      <p className="mt-2 text-[12px] leading-relaxed text-text-tertiary">{detail}</p>
    </div>
  );
}

function ConnectorArrow({ direction = 'right' }: { direction?: 'right' | 'left' | 'down' }) {
  const Icon = direction === 'left' ? ArrowLeft : direction === 'down' ? ArrowDown : ArrowRight;

  return (
    <div className="flex items-center justify-center text-line-cta">
      <Icon className="size-5" aria-hidden="true" />
    </div>
  );
}

export function ExtractQueryActDiagram() {
  return (
    <section
      className="not-prose my-5 overflow-hidden border border-line-structure bg-surface-bg shadow-[0_14px_36px_rgba(12,31,23,0.07)]"
      aria-label="Extract query act loop diagram"
    >
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-line-structure bg-surface-1 px-3 py-2">
        <div className="flex items-center gap-2">
          <span className="flex size-7 items-center justify-center border border-line-structure bg-surface-bg text-line-cta">
            <GitBranch className="size-4" aria-hidden="true" />
          </span>
          <span>
            <span className="block font-square text-[11px] font-semibold uppercase tracking-wider text-text-secondary">
              Local data loop
            </span>
            <span className="block text-[11px] text-text-tertiary">
              The same commands work for humans and agents.
            </span>
          </span>
        </div>
        <div className="flex items-center gap-1.5">
          <Badge variant="category">extract</Badge>
          <Badge variant="category">query</Badge>
          <Badge variant="category">act</Badge>
        </div>
      </div>

      <div className="grid gap-2 p-3 lg:grid-cols-[minmax(0,1fr)_2rem_minmax(0,1fr)_2rem_minmax(0,1fr)]">
        <DiagramNode
          icon={Cable}
          label="Any source"
          detail="Connector catalog metadata, credentials, and stream shape."
        />
        <ConnectorArrow />
        <DiagramNode
          icon={Database}
          label="Local warehouse"
          detail="pm etl run lands data locally for SQL without vendor infrastructure."
        />
        <ConnectorArrow />
        <DiagramNode
          icon={SearchCode}
          label="Decide"
          detail="pm query run answers the question with embedded DuckDB SQL."
        />
      </div>

      <div className="grid gap-2 border-t border-line-structure bg-surface-1 p-3 lg:grid-cols-[minmax(0,1fr)_2rem_minmax(0,1fr)_2rem_minmax(0,1fr)]">
        <div className="hidden lg:block" />
        <div className="hidden lg:block" />
        <DiagramNode
          icon={Bot}
          label="Plan write"
          detail="pm reverse plan / preview / run keeps write actions explicit."
        />
        <ConnectorArrow direction="left" />
        <DiagramNode
          icon={GitBranch}
          label="Any destination"
          detail="Approved reverse ETL writes results back to supported systems."
        />
        <div className="flex justify-center lg:hidden">
          <ConnectorArrow direction="down" />
        </div>
      </div>
    </section>
  );
}
