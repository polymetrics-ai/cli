# Polymetrics CLI SEO And AI Search Plan

Date: 2026-07-02

## Executive Summary

Goal: make Polymetrics CLI easy to understand, crawl, cite, and recommend for
queries around local-first ETL, reverse ETL CLI, embedded DuckDB SQL, connector
automation, and AI agent data workflows.

This plan does not assume we can guarantee a number-one ranking. Search and AI
answer visibility depends on indexing, relevance, freshness, competition,
third-party citations, and user behavior. The controllable work is to publish
clear, useful, technically crawlable, first-hand content that consistently
explains what Polymetrics is and why it is different.

Canonical positioning:

> Polymetrics CLI is a local-first data CLI for single-binary ETL, embedded
> DuckDB SQL, reverse ETL, connector automation, and AI agent data workflows.

Public tagline:

> One CLI to rule them all.

## Research Base

### Official Search Guidance

- Google Search Central's 2026 generative AI guidance says SEO fundamentals
  still matter because AI Overviews and AI Mode are rooted in core Search
  systems and use techniques such as retrieval-augmented generation and query
  fan-out.
- Google says the best long-term work is unique, valuable, non-commodity
  content with a clear technical structure, crawlability, semantic HTML where
  useful, good page experience, and reduced duplicate content.
- Google explicitly says `llms.txt` and special AI-only markup do not help or
  harm Google Search rankings. Keeping `llms.txt` can still help other tools,
  agents, and ecosystem consumers, but it is not a Google ranking shortcut.
- Google warns against creating pages primarily to manipulate generative AI
  answers, such as overproducing fan-out query pages, rewriting only for AI, or
  seeking inauthentic mentions.

Source:
[Google Search Central: Optimizing for generative AI features](https://developers.google.com/search/docs/fundamentals/ai-optimization-guide)

### README And Open Source Discovery

- GitHub's README guidance says a repository README should explain what the
  project does, why it is useful, how users get started, where to get help, and
  who maintains or contributes.
- GitHub also treats README, license, contributing guidance, and code of conduct
  as part of the repository's public contribution surface.
- Recent open-source documentation research frames README and CONTRIBUTING
  files as first-contact documents for contributors, but finds that many
  projects keep them minimal and procedural rather than community-building. For
  Polymetrics, the README should stay minimal while still doing positioning,
  onboarding, and contribution routing.

Sources:
[GitHub Docs: About READMEs](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes),
[The Introduction of README and CONTRIBUTING Files in Open Source Software Development](https://arxiv.org/abs/2502.18440)

### Latest AI Search Research

- "What Gets Cited: Competitive GEO in AI Answer Engines" (submitted May 25,
  2026) tests 252,000 citation trials across six LLMs. Its strongest practical
  result is that topical relevance and list/source position matter most; recent
  timestamps, explicit concrete information, completeness, and trust cues help;
  formatting-only changes have little impact.
- "Generative Engine Optimization: How to Dominate AI Search" argues that AI
  search can differ from traditional search by favoring earned third-party
  authoritative sources, machine-scannable justification, freshness, and
  engine-specific behavior. Treat this as useful directional evidence, not a
  substitute for official search guidance.
- "What Generative Search Engines Like and How to Optimize Web Content
  Cooperatively" introduces AutoGEO and shows that generative engines have
  learnable preferences when using retrieved content. This supports maintaining
  clear, evidence-rich, well-structured pages, but not manipulative rewrites.
- "Unveiling the Resilience of LLM-Enhanced Search Engines against Black-Hat
  SEO Manipulation" reports that LLM search engines resist most traditional
  black-hat SEO during retrieval, while newer LLM-specific spam patterns can
  increase manipulation. This is a reason to avoid spam tactics and build
  durable, evidence-backed content.

Sources:
[What Gets Cited: Competitive GEO in AI Answer Engines](https://arxiv.org/abs/2605.25517),
[Generative Engine Optimization: How to Dominate AI Search](https://arxiv.org/abs/2509.08919),
[What Generative Search Engines Like and How to Optimize Web Content Cooperatively](https://arxiv.org/abs/2510.11438),
[Unveiling the Resilience of LLM-Enhanced Search Engines against Black-Hat SEO Manipulation](https://arxiv.org/abs/2603.25500)

## What Changed In This Pass

- README rewritten around the new canonical line and "One CLI to rule them all."
- Old named competitor wording removed from the README, site metadata, and
  homepage FAQ.
- `/blog` added as a static, crawlable editorial surface.
- Three launch articles added:
  - `one-cli-to-rule-them-all`
  - `agent-native-data-workflows`
  - `local-first-data-engine`
- Article pages expose canonical metadata, Open Graph metadata, and
  `BlogPosting` JSON-LD.
- Homepage now links to the blog section.
- Navbar and footer now link to `/blog`.
- `sitemap.xml` route added for home, docs, connector pages, blog posts,
  changelog, and patterns.
- `robots.txt` route added with sitemap discovery.
- Existing `llms.txt` and `llms-full.txt` now include the blog content.

## Positioning Architecture

### Primary Query Cluster

Own these phrases naturally across README, homepage, docs, blog titles, and page
descriptions:

- local-first data CLI
- single-binary ETL
- ETL CLI
- reverse ETL CLI
- embedded DuckDB SQL
- AI agent data workflows
- connector automation
- approval-gated writes
- local encrypted credential vault

### Avoid

- Repeating keywords mechanically.
- Creating one thin page for every slight search variation.
- Claiming connector coverage beyond what is enabled in the current binary.
- Publishing generated content without first-hand details, commands, outputs,
  screenshots, or benchmarks.
- Treating `llms.txt` as an SEO hack for Google.

### Canonical Copy Blocks

Short:

> Polymetrics CLI is a local-first data CLI for ETL, embedded DuckDB SQL,
> reverse ETL, and AI-agent-safe automation.

Medium:

> Polymetrics CLI turns extract, query, and write-back automation into one local
> Go binary. Use it to pull connector data, query it with embedded DuckDB SQL,
> and run approval-gated reverse ETL from the same CLI contract humans and AI
> agents can operate.

Long:

> Polymetrics CLI is a local-first data engine for developers, data engineers,
> and AI agents. It combines connector-backed ETL, local warehouse storage,
> embedded DuckDB SQL, reverse ETL planning, approval-gated writes, schedules,
> and machine-readable JSON output in one command-line binary.

## Technical SEO Plan

### Implemented

- `metadataBase` set to `https://cli.polymetrics.ai`.
- Site title and Open Graph copy updated to the new positioning.
- Static `/blog` and `/blog/[slug]` pages added.
- Blog pages include `Blog` and `BlogPosting` structured data.
- `sitemap.ts` added.
- `robots.ts` added.
- Blog links added to homepage, navbar, footer, and `llms.txt`.

### Next

1. Verify production has:
   - `https://cli.polymetrics.ai/sitemap.xml`
   - `https://cli.polymetrics.ai/robots.txt`
   - `https://cli.polymetrics.ai/blog`
   - `https://cli.polymetrics.ai/llms.txt`
2. Submit sitemap in Google Search Console and Bing Webmaster Tools.
3. Add Search Console domain verification to deployment configuration.
4. Run Rich Results Test or Schema validator for the blog pages.
5. Add per-page canonical metadata to high-value docs pages if missing.
6. Add `lastModified` dates to docs and connector pages instead of using a
   single default date in the sitemap.
7. Add `og:image` assets for the homepage and blog articles.
8. Add performance budget checks for the website build.

## Content Plan

### Blog Series

Publish one substantial article per week for the first eight weeks. Each article
must include first-hand product details, command examples, tradeoffs, and a
clear conclusion.

Initial sequence:

1. One CLI To Rule Them All
2. Agent-Native Data Workflows Need Boring Contracts
3. Local-First Data Pipelines Without Warehouse Sprawl
4. Reverse ETL From A CLI: Plan, Preview, Approve, Run
5. DuckDB As The Default Local Analytics Engine
6. Why Connector Catalogs Need Tests, Not Just YAML
7. How To Run A GitHub Issues Data Loop With Polymetrics
8. Designing Exit Codes For AI Agents

### Connector Pages

Connector pages are a major search surface, but they must avoid becoming thin
catalog pages. Each enabled connector should include:

- What the connector reads and writes.
- Required credentials and scopes.
- Exact setup commands.
- A minimal working extract example.
- A query example.
- A reverse ETL example where writes are supported.
- Common failures and fixes.
- Rate-limit notes.
- Last verified date.
- Machine-readable `data.json` endpoint.

Planned connector pages can exist, but they should clearly say planned/native
port pending and should not imply production support.

### Docs Pages

High-value docs pages to improve next:

- `/docs/quickstart`
- `/docs/query`
- `/docs/reverse-etl`
- `/docs/agent-guide`
- `/docs/connectors`
- `/docs/scheduling`
- `/docs/architecture`

Each should have:

- One H1.
- A clear description.
- A runnable command block.
- Expected JSON output.
- Troubleshooting section.
- Last updated date.
- Links to related docs and blog posts.

## AI Search And Citation Plan

AI answer engines need sources they can retrieve and quote accurately. The best
Polymetrics pages should be citation-ready:

- Specific rather than vague.
- Freshly dated.
- Written from first-hand product knowledge.
- Clear about limitations.
- Backed by commands, output, screenshots, or benchmark data.
- Easy to skim with headings and tables.
- Not overpromotional.

Actions:

1. Keep `llms.txt` and `llms-full.txt` accurate for tools that use them.
2. Maintain structured metadata and readable HTML first; do not rely on
   Markdown-only AI files.
3. Publish small reproducible datasets and command outputs for blog examples.
4. Add a "facts" section to major pages with stable claims:
   - binary name
   - license
   - current connector catalog count
   - enabled connector count
   - supported sync modes
   - supported operating systems
5. Seek authentic third-party references:
   - launch posts
   - engineering forums
   - relevant open-source directories
   - connector ecosystem discussions
   - technical podcasts/newsletters only when there is a real story

## Measurement Plan

### Weekly

- Google Search Console indexed pages.
- Query impressions and clicks for primary terms.
- Crawl errors and sitemap coverage.
- Top landing pages and page experience metrics.
- GitHub traffic to README and docs links.

### Biweekly

Manual AI-search checks:

- "local-first ETL CLI"
- "reverse ETL CLI"
- "DuckDB ETL CLI"
- "AI agent data workflow CLI"
- "Polymetrics CLI"
- "one CLI to rule them all data"

Record:

- Whether Polymetrics is mentioned.
- Which URL is cited.
- Whether the description is accurate.
- Which competing categories appear.
- What source gaps are visible.

### Monthly

- Update README and homepage claims if connector/support status changes.
- Refresh top docs pages with verified command output.
- Review thin pages and merge or improve weak content.
- Check backlinks and authentic mentions.

## 30-60-90 Day Roadmap

### First 30 Days

- Verify sitemap and robots in production.
- Submit sitemap to Search Console and Bing.
- Publish the first four blog posts.
- Add per-page last updated dates to docs.
- Add Open Graph images.
- Add canonical metadata to high-value docs.
- Add expected output blocks to quickstart, query, reverse ETL, and agent docs.

### 60 Days

- Improve enabled connector pages with real setup and troubleshooting details.
- Add a benchmark page for local extract/query performance.
- Add a "compare categories" doc that compares deployment model, data locality,
  and agent contract without naming or attacking specific vendors.
- Add real screenshots or terminal recordings for homepage and README.
- Create a recurring docs freshness check in CI.

### 90 Days

- Publish a connector contribution guide with a complete worked example.
- Publish case-study style posts using open-source datasets and repositories.
- Add docs for Homebrew once the tap exists.
- Build a search analytics dashboard from Search Console exports.
- Run a content audit: remove, merge, or enrich pages that do not earn
  impressions, links, or user engagement.

## Acceptance Criteria

- Production `sitemap.xml` includes home, docs, connectors, blog, changelog, and
  patterns.
- Production `robots.txt` points to the sitemap.
- `/blog` and all article pages render without client-side data fetching.
- Blog article source contains `BlogPosting` JSON-LD.
- README contains no old named competitor wording.
- README explains what the project does, why it is useful, how to start, where
  docs live, and how to contribute.
- Google Search Console shows submitted sitemap discovered and processed.
- Search result snippets for branded queries describe Polymetrics as a
  local-first data CLI.

## Risks

- Overstating enabled connector support can damage trust. Keep catalog and
  enabled support separate.
- Publishing generated pages without useful setup details can look like thin
  content.
- Chasing AI-search hacks can conflict with Google guidance and recent research
  on black-hat LLM search manipulation.
- A beautiful README that is too long may hide the quickstart. Keep the README
  minimal and push deep detail into docs.
- Third-party AI engines may cite stale pages. Add dates and keep high-value
  pages current.

## Operating Rule

Every public page should answer four questions quickly:

1. What is Polymetrics?
2. Why should I trust it?
3. How do I run it?
4. What can I cite or copy from this page?
