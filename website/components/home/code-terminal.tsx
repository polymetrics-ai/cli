import { type ReactNode } from 'react';
import { CornerBox } from '@/components/ui/corner-box';

/* ── Design tokens (matching globals.css) ──────────────────────────────── */
const C = {
  comment:    '#8c8c84',
  cmdPrimary: '#222220',
  subcommand: '#404039',
  flag:       '#5b6b3a',
  flagValue:  '#3a3a35',
  jsonOutput: '#6b6b66',
  backslash:  '#a7a7a0',
  equals:     '#a7a7a0',
  sqlString:  '#6b6b66',
  sqlKeyword: '#404039',
  normal:     'var(--text-secondary)',
} as const;

/* ── Helpers ───────────────────────────────────────────────────────────── */
function span(key: string, color: string, text: string, extra?: React.CSSProperties): ReactNode {
  return (
    <span key={key} style={{ color, ...extra }}>
      {text}
    </span>
  );
}

const FLAGS = [
  '--connector', '--from-env', '--source', '--destination',
  '--stream', '--primary-key', '--cursor', '--table',
  '--connection', '--sql', '--json', '-p',
];

const SUBCOMMANDS = [
  'install', 'init', 'credentials', 'connections', 'etl',
  'query', 'run', 'add', 'create',
];

const SQL_KEYWORDS = ['SELECT', 'COUNT', 'FROM', 'WHERE', 'GROUP BY'];

/* ── Per-line tokenizer ────────────────────────────────────────────────── */
function tokenizeLine(line: string, index: number): ReactNode {
  // Empty line
  if (line === '') return <span key={index}>{'\n'}</span>;

  // Comment line (including the JSON output comment)
  if (line.trimStart().startsWith('#')) {
    const isJsonOutput = line.includes('→ {');
    if (isJsonOutput) {
      // Split on the arrow so the comment marker stays muted
      const arrowIdx = line.indexOf('→');
      const prefix   = line.slice(0, arrowIdx);     // "# "
      const arrow    = '→ ';
      const json     = line.slice(arrowIdx + 2);    // {"kind":…}
      return (
        <span key={index} style={{ color: C.comment }}>
          {prefix}
          <span style={{ color: C.jsonOutput, fontStyle: 'italic' }}>
            {arrow}{json}
          </span>
          {'\n'}
        </span>
      );
    }
    return (
      <span key={index} style={{ color: C.comment }}>
        {line}{'\n'}
      </span>
    );
  }

  // All other lines: tokenize token-by-token
  const nodes: ReactNode[] = [];
  let remaining = line;
  const key = () => `${index}-${nodes.length}`;

  while (remaining.length > 0) {
    // Trailing backslash continuation
    if (remaining === '\\') {
      nodes.push(span(key(), C.backslash, '\\'));
      remaining = '';
      break;
    }

    // Flags (--foo or -p)
    const flagMatch = FLAGS
      .map(f => ({ flag: f, match: remaining.startsWith(f) ? f : null }))
      .find(x => x.match !== null);
    if (flagMatch) {
      const f = flagMatch.flag;
      nodes.push(span(key(), C.flag, f));
      remaining = remaining.slice(f.length);

      // What follows a flag: optional space + value
      // Value ends at the next space or backslash or end
      const afterFlag = remaining.match(/^( )(.*)/);
      if (afterFlag) {
        nodes.push(span(key(), C.normal, ' '));
        remaining = afterFlag[2];

        // If the next thing is another flag, don't consume it as a value
        const nextIsFlag = FLAGS.some(nf => remaining.startsWith(nf));
        const nextIsBackslash = remaining.startsWith('\\');

        if (!nextIsFlag && !nextIsBackslash && remaining.length > 0) {
          // Grab up to next space or backslash; that's the value
          const valMatch = remaining.match(/^([^\s\\]+)(.*)/);
          if (valMatch) {
            let val = valMatch[1];
            const rest = valMatch[2];

            // --sql has a quoted SQL string value
            if (f === '--sql') {
              // val starts with a quote; grab until closing quote (may span remaining)
              if (val.startsWith('"') || val.startsWith("'")) {
                const q = val[0];
                // Consume until closing quote in full remaining
                const closeIdx = remaining.indexOf(q, 1);
                if (closeIdx !== -1) {
                  val = remaining.slice(0, closeIdx + 1);
                  remaining = remaining.slice(closeIdx + 1);
                  nodes.push(tokenizeSql(val, key()));
                  continue;
                }
              }
            }

            nodes.push(span(key(), C.flagValue, val));
            remaining = rest;
          }
        }
      }
      continue;
    }

    // Leading spaces (indentation)
    const spaceMatch = remaining.match(/^( +)(.*)/);
    if (spaceMatch) {
      nodes.push(<span key={`sp-${remaining}`}>{spaceMatch[1]}</span>);
      remaining = spaceMatch[2];
      continue;
    }

    // Command token: first word of a non-indented line
    // Already consumed indent above, so check if nodes is empty (first token)
    const wordMatch = remaining.match(/^([^\s\\]+)(.*)/);
    if (wordMatch) {
      const word = wordMatch[1];
      remaining  = wordMatch[2];

      if (word === 'go' || word === 'pm') {
        nodes.push(span(key(), C.cmdPrimary, word, { fontWeight: 700 }));
      } else if (SUBCOMMANDS.includes(word)) {
        nodes.push(span(key(), C.subcommand, word));
      } else {
        // plain argument / positional value
        nodes.push(span(key(), C.normal, word));
      }
      continue;
    }

    // Fallback: emit one char and advance
    nodes.push(<span key={`fc-${remaining}`}>{remaining[0]}</span>);
    remaining = remaining.slice(1);
  }

  return (
    <span key={index}>
      {nodes}
      {'\n'}
    </span>
  );
}

/* Tokenize an SQL string literal (the whole quoted value) */
function tokenizeSql(raw: string, keyPrefix: string): ReactNode {
  // raw = "SELECT assignee, COUNT(*) AS open FROM issues WHERE state='open' GROUP BY 1"
  // Opening and closing double-quote belong to the value
  const inner = raw.slice(1, raw.length - 1); // strip surrounding double quotes

  const nodes: ReactNode[] = [];
  const key = () => `${keyPrefix}-${nodes.length}`;
  nodes.push(span(key(), C.flagValue, '"'));

  // Tokenize inside: find SQL keywords and single-quoted strings
  let s = inner;
  while (s.length > 0) {
    // SQL keyword match (case-sensitive as per QUICKSTART)
    const kwMatch = SQL_KEYWORDS.find(kw => s.startsWith(kw));
    if (kwMatch) {
      nodes.push(span(key(), C.sqlKeyword, kwMatch, { fontWeight: 600 }));
      s = s.slice(kwMatch.length);
      continue;
    }

    // Single-quoted string
    if (s.startsWith("'")) {
      const closeIdx = s.indexOf("'", 1);
      if (closeIdx !== -1) {
        nodes.push(span(key(), C.sqlString, s.slice(0, closeIdx + 1)));
        s = s.slice(closeIdx + 1);
        continue;
      }
    }

    // Plain char
    nodes.push(
      <span key={key()} style={{ color: C.normal }}>
        {s[0]}
      </span>,
    );
    s = s.slice(1);
  }

  nodes.push(span(key(), C.flagValue, '"'));
  return <span key={keyPrefix}>{nodes}</span>;
}

/* ── QUICKSTART content ────────────────────────────────────────────────── */
const QUICKSTART = `# Install (Go 1.24+)
go install polymetrics.ai/cmd/pm@latest

pm init
pm credentials add github \\
  --connector github --from-env token=GITHUB_TOKEN

pm connections create my-github \\
  --source github:github --destination warehouse:warehouse \\
  --stream issues --primary-key id \\
  --cursor updated_at --table issues

pm etl run --connection my-github --stream issues --json
# → {"kind":"ETLRun","run":{"records_read":142,"records_loaded":142}}

pm query run --table issues \\
  --sql "SELECT assignee, COUNT(*) AS open FROM issues WHERE state='open' GROUP BY 1" \\
  --json`;

/* ── Component ─────────────────────────────────────────────────────────── */
export function CodeTerminal() {
  const lines = QUICKSTART.split('\n');

  return (
    <div className="terminal-3d">
      <CornerBox className="overflow-hidden p-0">
        {/* Traffic-light chrome bar */}
        <div className="border-b border-line-structure bg-surface-1 px-4 py-2.5 flex items-center gap-1.5">
          <span className="h-3 w-3 rounded-full bg-red-400/80" />
          <span className="h-3 w-3 rounded-full bg-yellow-400/80" />
          <span className="h-3 w-3 rounded-full bg-green-400/80" />
          <span className="ml-2 font-mono text-[11px] text-text-tertiary">pm · quickstart</span>
        </div>

        {/* Highlighted code */}
        <pre className="overflow-x-auto p-5 font-mono text-[12px] leading-relaxed bg-surface-bg">
          <code>
            {lines.map((line, i) => tokenizeLine(line, i))}
          </code>
        </pre>
      </CornerBox>
    </div>
  );
}
