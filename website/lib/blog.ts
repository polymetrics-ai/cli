export type BlogSection = {
  heading: string;
  body: string[];
  points?: string[];
  code?: string;
  evidenceIds?: string[];
};

export type BlogEvidenceKind = 'pull_request' | 'commit' | 'workflow';

export type BlogEvidence = {
  id: string;
  kind: BlogEvidenceKind;
  label: string;
  url: string;
  apiUrl: string;
  openLabel: string;
  snapshot: {
    title: string;
    status: string;
    ref: string;
    verifiedAt: string;
    summary: string;
    stats: Array<{ label: string; value: string }>;
  };
};

export type BlogPost = {
  slug: string;
  title: string;
  description: string;
  publishedAt: string;
  updatedAt: string;
  readingTime: string;
  category: string;
  tags: string[];
  summary: string;
  repositoryCta?: {
    label: string;
    href: string;
  };
  evidence?: BlogEvidence[];
  sections: BlogSection[];
};

export const BLOG_POSTS: BlogPost[] = [
  {
    slug: 'human-harnesses',
    title: 'Humans Need Harnesses Too',
    description:
      'One pull request crossed 1.9 million changed lines and turned merging into archaeology. The recovery became the harness architecture we now use for code, data, and humans.',
    publishedAt: '2026-07-16',
    updatedAt: '2026-07-16',
    readingTime: '14 min read',
    category: 'Build in public',
    tags: ['human harnesses', 'GitHub Actions', 'approval gates', 'AI agents'],
    summary:
      'A giant pull request turned GitHub into a loading spinner and taught us that intent, evidence, approval, and mutation must stay visible for every operator, including the human at the keyboard.',
    repositoryCta: {
      label: 'Star Polymetrics on GitHub',
      href: 'https://github.com/polymetrics-ai/cli',
    },
    evidence: [
      {
        id: 'precursor-pr-27',
        kind: 'pull_request',
        label: 'PR #27',
        url: 'https://github.com/polymetrics-ai/cli/pull/27',
        apiUrl: 'https://api.github.com/repos/polymetrics-ai/cli/pulls/27',
        openLabel: 'Open PR #27 on GitHub',
        snapshot: {
          title: 'Connector architecture v2 bundles and certification harness',
          status: 'Closed, not merged',
          ref: '5ce0e094f7e937859e344f37782b028f4bd1fcbf',
          verifiedAt: '2026-07-16',
          summary:
            'The original migration delivery unit. It accumulated 1,961,878 changed lines across 24,803 files before closing without a merge.',
          stats: [
            { label: 'Changed lines', value: '1,961,878' },
            { label: 'Additions', value: '1,932,552' },
            { label: 'Deletions', value: '29,326' },
            { label: 'Files', value: '24,803' },
          ],
        },
      },
      {
        id: 'migration-pr-29',
        kind: 'pull_request',
        label: 'PR #29',
        url: 'https://github.com/polymetrics-ai/cli/pull/29',
        apiUrl: 'https://api.github.com/repos/polymetrics-ai/cli/pulls/29',
        openLabel: 'Open PR #29 on GitHub',
        snapshot: {
          title: 'Complete connector architecture v2 migration',
          status: 'Merged',
          ref: 'a1c1af7535f9ac273f701a40b2a32d7d376ea18a',
          verifiedAt: '2026-07-16',
          summary:
            'The conflict-resolved replacement merged 2,792,444 changed lines across 28,733 files on July 6, 2026.',
          stats: [
            { label: 'Changed lines', value: '2,792,444' },
            { label: 'Additions', value: '2,242,501' },
            { label: 'Deletions', value: '549,943' },
            { label: 'Files', value: '28,733' },
          ],
        },
      },
      {
        id: 'migration-merge-commit',
        kind: 'commit',
        label: 'Commit 605b006',
        url: 'https://github.com/polymetrics-ai/cli/commit/605b006e5aa1adae697d5b7dd26ec485c570c250',
        apiUrl:
          'https://api.github.com/repos/polymetrics-ai/cli/commits/605b006e5aa1adae697d5b7dd26ec485c570c250',
        openLabel: 'Open commit 605b006 on GitHub',
        snapshot: {
          title: 'Complete connector architecture v2 migration (#29)',
          status: 'Committed to main',
          ref: '605b006e5aa1adae697d5b7dd26ec485c570c250',
          verifiedAt: '2026-07-16',
          summary:
            'The exact main-branch commit produced by the recovery merge. This is the durable object behind the story, not a reconstructed screenshot.',
          stats: [
            { label: 'Changed lines', value: '2,792,444' },
            { label: 'Additions', value: '2,242,501' },
            { label: 'Deletions', value: '549,943' },
            { label: 'SHA', value: '605b006' },
          ],
        },
      },
      {
        id: 'issue-first-pr-47',
        kind: 'pull_request',
        label: 'PR #47',
        url: 'https://github.com/polymetrics-ai/cli/pull/47',
        apiUrl: 'https://api.github.com/repos/polymetrics-ai/cli/pulls/47',
        openLabel: 'Open PR #47 on GitHub',
        snapshot: {
          title: 'Add issue-first delivery system',
          status: 'Merged',
          ref: '7eb144f59b548dcab4831bbf9df14cea45775985',
          verifiedAt: '2026-07-16',
          summary:
            'Introduced the issue-scoped delivery contracts and evidence structure that followed the oversized migration.',
          stats: [
            { label: 'Changed lines', value: '3,164' },
            { label: 'Additions', value: '2,931' },
            { label: 'Deletions', value: '233' },
            { label: 'Files', value: '50' },
          ],
        },
      },
      {
        id: 'parent-orchestrator-pr-51',
        kind: 'pull_request',
        label: 'PR #51',
        url: 'https://github.com/polymetrics-ai/cli/pull/51',
        apiUrl: 'https://api.github.com/repos/polymetrics-ai/cli/pulls/51',
        openLabel: 'Open PR #51 on GitHub',
        snapshot: {
          title: 'Add parent issue orchestrator',
          status: 'Merged',
          ref: '484c3dcad57e8e5c099e1ad1e85438006f39167a',
          verifiedAt: '2026-07-16',
          summary:
            'Added explicit parent ownership for dependency graphs, sub-PR integration, review coverage, and the final human gate.',
          stats: [
            { label: 'Changed lines', value: '1,435' },
            { label: 'Additions', value: '1,316' },
            { label: 'Deletions', value: '119' },
            { label: 'Files', value: '31' },
          ],
        },
      },
      ...[
        {
          id: 'issue-guard-workflow',
          label: 'PR issue guard',
          file: 'pr-issue-guard.yml',
          sha: '8640b533d857581f946d40163c5bc9804eaf9d57',
          size: '962 B',
          summary: 'Checks the accepted issue-reference shape in pull-request titles and bodies.',
        },
        {
          id: 'conventions-workflow',
          label: 'Conventions',
          file: 'conventions.yml',
          sha: '23c69d3df3609fa7951a2d7b09fd77c56b2df590',
          size: '2,041 B',
          summary: 'Checks branch naming and Conventional Commit pull-request titles.',
        },
        {
          id: 'gsd-workflow',
          label: 'GSD evidence',
          file: 'gsd-workflow.yml',
          sha: 'a14de8d214ca468111cd795af09edd66487e849b',
          size: '549 B',
          summary: 'Requires planning and test evidence when production Go code changes.',
        },
        {
          id: 'verify-workflow',
          label: 'Verify workflow',
          file: 'verify.yml',
          sha: '6ca4873f168083e16f8c6117e61b6df4f84fb46a',
          size: '858 B',
          summary: 'Runs the repository verification gate and rejects generated-file drift.',
        },
        {
          id: 'security-workflow',
          label: 'Security workflow',
          file: 'security.yml',
          sha: '30e31ffa0a423563895443c59bfb992acfb24205',
          size: '1,650 B',
          summary: 'Runs the repository vulnerability and static-analysis security checks.',
        },
        {
          id: 'website-workflow',
          label: 'Website workflow',
          file: 'website.yml',
          sha: '79af5f1f4371832468724974d79a07be43fa445f',
          size: '5,286 B',
          summary: 'Builds and tests the website, publishes its image, and controls deployment.',
        },
        {
          id: 'release-workflow',
          label: 'Release workflow',
          file: 'release.yml',
          sha: '20a5ea9b3d0b6b09b5ecedf2db542a63a7692f1b',
          size: '1,427 B',
          summary: 'Separates main-branch integration from version and artifact publication.',
        },
      ].map(({ id, label, file, sha, size, summary }): BlogEvidence => ({
        id,
        kind: 'workflow',
        label,
        url: `https://github.com/polymetrics-ai/cli/blob/main/.github/workflows/${file}`,
        apiUrl: `https://api.github.com/repos/polymetrics-ai/cli/contents/.github/workflows/${file}?ref=main`,
        openLabel: `Open ${file} on GitHub`,
        snapshot: {
          title: file,
          status: 'Tracked on main',
          ref: sha,
          verifiedAt: '2026-07-16',
          summary,
          stats: [
            { label: 'Branch', value: 'main' },
            { label: 'File size', value: size },
            { label: 'Blob', value: sha.slice(0, 7) },
          ],
        },
      })),
    ],
    sections: [
      {
        heading: 'The PR that ate the repository',
        body: [
          'This story starts with PR #27 and 1,961,878 changed lines. That is not a motivational metaphor. It was connector definitions, schemas, fixtures, generated documentation, tests, and enough JSON to make the diff viewer reconsider its career choices. That PR closed without merging; the conflict-resolved PR #29 eventually landed 2,792,444 changed lines in commit 605b006.',
          'At that size, code review stops being code review. The scrollbar becomes a rounding error. A useful comment on line 412,000 is less a review note and more a message in a bottle. Every merge conflict asks the same cheerful question: do you still remember why this file changed three weeks ago?',
          'The individual changes were mostly reasonable. The package was the disaster. One PR was trying to be the roadmap, the work queue, the integration branch, the test report, and the release decision at the same time. When one check failed, the answer was not which slice broke. The answer was yes.',
          'I eventually realized the agents were not the main problem. I had given fast workers one enormous room, no lanes, and a single door marked MERGE. The million-line PR did not need a braver reviewer. It needed an architecture that made work smaller before review began.',
        ],
        evidenceIds: ['precursor-pr-27', 'migration-pr-29', 'migration-merge-commit'],
      },
      {
        heading: 'The tool after the fire',
        body: [
          'I wanted a real CLI for every connector, the way gh is my daily driver for GitHub. Not a configuration file interpreted by a cloud service I cannot see, but a command surface that can inspect a connector, extract data, query it locally, plan a write, and explain every failure.',
          'I had been using the large AI coding platforms to chase that goal. They could produce code quickly, but I was spending too much time arranging agents, recovering context, and discovering that five fast workers editing the same generated file is just a merge conflict speedrun with nicer typography.',
          'Then I found Pi. It was small, direct, and easy to shape around the repository. Combined with GSD, it gave me a loop I could inspect: define the issue, plan the slice, write the failing test, make it pass, verify the result, and leave evidence for the next person or agent. Boring steps, in the best possible sense.',
          'The connector rewrite now stores each integration as a JSON bundle and interprets it through a shared engine. API operations are classified as streams, direct reads, writes, binary transfers, native protocol work, hooks, or typed exclusions. That classification matters because a list operation, a file download, and a DELETE should not inherit the same runtime policy.',
          'That local loop eventually raised a more awkward question: who checks whether the loop itself is behaving? I am leaving that unanswered here, because it became its own engineering story.',
        ],
      },
      {
        heading: 'What the repository actually contains',
        body: [
          'The giant PR did not become giant because one feature got carried away. It became giant because the surface area is genuinely large, and I had confused large scope with large delivery units.',
          'These are repository inventory numbers, not a claim that every connector is production-certified. At the time of writing there are 547 connector definition directories, each with an API-surface inventory. Those files contain 29,129 classified endpoint entries: 14,780 GET reads, 3 HEAD checks, 14,169 explicit POST, PUT, PATCH, or DELETE mutations, and 177 hook, wildcard, GraphQL, WebSocket, or composite-method rows. The explicit mutations include 2,903 DELETE operations.',
          'The same bundles define 7,088 ETL streams. Some operations are implemented, some are deliberately excluded, some depend on hooks or native code, and some still need live certification. The inventory tells us the size and shape of the work. It does not let us skip the proof that a connector behaves correctly against a real service.',
        ],
        points: [
          '547 connector bundles with explicit API-surface inventories.',
          '29,129 operations classified before they are exposed as product behavior.',
          '14,780 GET reads, 3 HEAD checks, and 14,169 explicit HTTP mutations.',
          '177 mixed or nonstandard method rows remain visible instead of being mislabeled as writes.',
          '2,903 DELETE operations are called out explicitly inside the mutation inventory.',
          '7,088 ETL streams defined for conformance and certification work.',
        ],
      },
      {
        heading: 'The uncomfortable realization',
        body: [
          'From the start, Polymetrics had an agent contract: machine-readable JSON, stable exit behavior, credentials referenced instead of printed, and reverse ETL split into plan, preview, approval, and execution. The point was to let an LLM use the same CLI as a person without giving it a quiet path to mutate a destination.',
          'Then one night I watched myself work. It was late, I was tired, and I was about to aim a write at the wrong environment. The thing that saved me was not better judgment. It was the same stop I had designed for an agent.',
          'That is the uncomfortable part: I am also an unreliable agent, just one with coffee and commit access. Humans paste a correct command into the wrong terminal. We skim diffs, approve familiar-looking output, reuse stale plans, and reach for force when the system slows us down. A safety model that protects data only when an AI is operating it has misunderstood the operator.',
        ],
      },
      {
        heading: 'The runtime harness',
        body: [
          'A harness turns a mutation from one action into a sequence with named states. First describe the intended change. Then calculate and preview the concrete effect. Bind approval to that plan. Only then execute it. A failure returns structured status and leaves the destination unchanged whenever the operation boundary allows it.',
          'The August 4 target is to make that model explicit for people as well as agents. Destructive and sensitive paths should be unavailable until policy enables them. Approval should be scoped to a specific plan rather than becoming an ambient yes to everything. The same command should behave predictably from a terminal, CI job, cron entry, or agent loop.',
          'This is also where honesty matters. A mapped operation is not automatically an executable command, and an executable command is not automatically certified. The inventory, runtime policy, conformance tests, and live certification are separate gates because each answers a different question.',
        ],
        code: `pm reverse plan candidates_to_github --source-table candidates --destination github:github-local --action create_issue --map title:title
pm reverse preview <plan-id> --json
pm reverse run <plan-id> --approve <approval-token> --json`,
      },
      {
        heading: 'The repository became a harness',
        body: [
          'The million-line PR gave us a surprisingly useful requirements document. First, one parent issue owns the outcome and dependency graph. Then bounded sub-issues own one behavior each. Every worker gets an isolated worktree and an explicit file boundary, because asking several agents to share one checkout is not collaboration; it is competitive editing.',
          'Each sub-issue produces a stacked PR into the parent branch. That PR carries its own red and green evidence, focused checks, and review coverage. The parent branch is where the slices meet, full verification runs again, and cross-slice contradictions finally have somewhere obvious to appear.',
          'The parent PR into main stays human-gated. A sub-PR can prove that one component works; it cannot declare that thirteen individually green components form a coherent release. The last approval belongs to somebody looking at the assembled product, preferably before 2am.',
          'That is the structural harness: intent before code, bounded work before parallelism, evidence before integration, review before release, and an immutable artifact before deployment. GitHub Actions executes much of it, but the architecture begins before the first diff and ends after rollout health.',
        ],
        code: `parent issue -> dependency graph -> bounded sub-issues
sub-issue   -> isolated worktree -> red/green -> stacked PR
stacked PR  -> checks + review -> parent branch
parent      -> full verification -> human merge -> release`,
        evidenceIds: ['issue-first-pr-47', 'parent-orchestrator-pr-51'],
      },
      {
        heading: 'Intent before diff',
        body: [
          'Every non-trivial change starts with an issue that states the objective, scope, exclusions, acceptance criteria, verification, safety notes, and review route. This is less exciting than opening an editor and typing furiously, but so is wearing a seat belt. The excitement was never the useful part.',
          'The PR Issue Guard checks that the title and PR body use an accepted issue-reference shape. It does not prove that the issue exists or that the diff matches its scope; reviewers still own that judgment. The conventions workflow checks the branch name and Conventional Commit PR title, so the change identifies itself before anyone reads the implementation.',
          'When production Go code under cmd or internal changes, the GSD workflow checks for planning and test evidence. In practice that means a plan, a TDD ledger, and a verification checklist exist before production edits. The ledger does not prove the code is correct, but it makes a useful distinction visible: did a test fail because the capability was absent, and did the same test pass because of the change?',
          'This is the code equivalent of binding an approval to a plan. The issue describes the intended mutation; the branch and PR limit its scope; the test demonstrates the behavioral delta. A reviewer can challenge any of those layers instead of reverse-engineering intent from the final diff.',
        ],
        evidenceIds: ['issue-guard-workflow', 'conventions-workflow', 'gsd-workflow'],
      },
      {
        heading: 'Proof before confidence',
        body: [
          'The general verification workflow installs the pinned Go toolchain and linter, runs make verify, and then fails if verification changed generated files. A green check is a receipt, not a mood. That distinction becomes important after the third hour of a merge when everybody feels extremely confident and remembers almost nothing.',
          'The generated-file check is easy to overlook. A generator that produces an uncommitted diff is evidence that source and published artifacts disagree, so the harness treats drift as a failure rather than a cleanup task.',
          'Security runs in parallel: govulncheck inspects reachable Go vulnerabilities, CodeQL analyzes Go and JavaScript or TypeScript, and dependency review rejects newly introduced high-severity dependency risk. A weekly OpenSSF Scorecard run adds a slower repository-level view of supply-chain posture.',
          'The website has its own integration harness. It starts PostgreSQL 17, regenerates website data and checks for drift, typechecks, runs unit tests, installs Chromium, exercises the site with Playwright, creates a production build, and then builds the container image. That sequence catches failures that a component test cannot: migrations, generated catalog mismatches, route rendering, browser behavior, and production compilation.',
        ],
        points: [
          'Deterministic checks: formatting, linting, tests, builds, and generated-file drift.',
          'Security checks: govulncheck, CodeQL, dependency review, and scheduled Scorecard analysis.',
          'Website checks: PostgreSQL migrations, generated data, typecheck, unit tests, Chromium e2e, build, and image construction.',
        ],
        evidenceIds: ['verify-workflow', 'security-workflow', 'website-workflow'],
      },
      {
        heading: 'Review, fix, repeat, locally',
        body: [
          'Review used to arrive later, through a remote pull-request bot. That produced useful findings, but it also put credentials, network availability, provider quotas, and GitHub event timing on the critical path. A correction loop should not need to leave the machine just to discover that the fix introduced another bug.',
          'The default gate now runs locally inside the Pi delivery loop. After implementation passes verification, the orchestrator records the exact candidate head SHA and base SHA, then launches an independent read-only reviewer. The reviewer can inspect and search the scoped code, but it cannot edit files, push commits, post comments, merge branches, or quietly turn its own opinion into a mutation.',
          'Every actionable finding gets a disposition: accept it, accept it with a narrower change, decline it with a reason, defer it to a named follow-up, or stop for a human. Accepted work goes to an isolated repair worker with a bounded write scope. That worker fixes the code and reruns focused tests; verification runs again; then the reviewer sees the new exact head. The loop can repeat for up to four correction rounds before it stops pretending persistence is wisdom and asks a person.',
          'The live Pi session that prompted this update made the loop wonderfully unglamorous. A review sidecar returned three findings. One was accepted with modification, one was declined with a recorded scope reason, and one was accepted. The isolated repair worker fixed the accepted items and made the full verification gate green. Verification then caught trailing whitespace in the worker evidence and sent the same bounded worktree around once more. That second lap is the feature, not an embarrassment.',
          'Remote PR-bot review can still run as supplemental shadow or canary evidence during the cutover. It is no longer the default delivery gate. GitHub receives the branch, PR, checks, and durable review summary; the fast review-and-repair conversation happens locally, where a moved head immediately invalidates the old verdict.',
        ],
        code: `implement -> verify -> review exact head -> disposition
                      ^                         |
                      |                         v
                 re-review <- verify <- isolated fix`,
      },
      {
        heading: 'Release and deployment are mutations too',
        body: [
          'Merging code is not the final write. The release workflow lets release-please assemble the version and changelog on main, and GoReleaser builds artifacts only when a release is actually created. That separates ordinary integration from publication.',
          'The website follows a similar boundary. Pull requests can build the image, but only a main-branch push or an explicit dispatch can publish it. Deployment requires the website deploy variable, uses a self-hosted Tailscale runner, passes the image tagged with the exact commit SHA to the Quadlet deployment script, and verifies rollout health. The deploy consumes an immutable input instead of rebuilding whatever happens to be on the server.',
          'This is the production form of plan and execute: CI proves one commit, the registry stores an image for that commit, and the deploy step rolls out that exact image. If those identities diverge, the harness is no longer describing the mutation it performs.',
        ],
        evidenceIds: ['release-workflow', 'website-workflow'],
      },
      {
        heading: 'What GitHub really blocks',
        body: [
          'There is a difference between a workflow that runs and a check that GitHub requires before main can move. GitHub does not enforce architectural intentions, good vibes, or the sentence trust me in a PR comment. It enforces configured rules.',
          'At the time of writing, main uses strict required status checks for verify, govulncheck, CodeQL, branch-name, and pr-title, and the rule applies to administrators. It also requires linear history and resolved review conversations. A branch that is behind main must be brought current before those checks can authorize the merge.',
          'Other useful workflows also run, including the issue guard, GSD evidence check, dependency review, and website suite when relevant files change. They are part of our delivery contract, but they are not all listed as required branch-protection contexts today. Saying otherwise would turn a process expectation into a false technical guarantee.',
          'The parent PR is still human-gated. Sub-PRs can be integrated into a parent branch after scoped checks and review coverage, but the parent branch does not merge to main automatically. That gate matters because a green collection of local changes can still be incoherent as one product release.',
        ],
      },
      {
        heading: 'What the harness still does not do',
        body: [
          'Now for the paragraph marketing pages usually hide behind a gradient: the repository is active and the harness is not finished. Branch protection currently requires status checks but not a positive review count. The human gate on the parent PR is repository policy, not a GitHub rule that mathematically prevents every maintainer from merging without review.',
          'Local automated review is still input, not an oracle. Its verdict binds the candidate and base SHAs, the policy bundle, the observed reviewer model and thinking level, independent checks, findings, dispositions, and creation time. Move the head and the evidence expires. Keep the head fixed and the reviewer still cannot approve its own edits, because it has no edit path; a separate bounded worker must make every accepted change.',
          'The production environment does not yet require an environment reviewer, so a qualifying main push can deploy when the deployment variable is enabled. Actions are version-tagged rather than pinned to full commit SHAs, and repository settings do not currently enforce SHA pinning. Those are real hardening opportunities, and naming them is part of the harness: an unrecorded gap is impossible to prioritize.',
        ],
        points: [
          'Add an enforced review requirement or preserve an auditable human-gate record for parent merges.',
          'Add a production environment approval when deployment independence becomes more important than immediate rollout.',
          'Pin third-party actions by commit and enable repository-level pinning policy.',
          'Invalidate local review evidence when the head moves, and stop for a human when the correction cap is reached.',
        ],
      },
      {
        heading: 'Same contract, both species',
        body: [
          'The useful idea is not that humans and agents are identical. It is that neither should receive a special path around the state machine. A shell script should not get an easier write path than an agent. A maintainer should not get a less reproducible deployment because they know the server. Familiarity is not evidence.',
          'Stable JSON and exit behavior make runtime state legible. Issues, TDD ledgers, checks, and reviewed commit ranges make repository state legible. Immutable image tags make deployment state legible. In each case the harness replaces an assumption with an object we can inspect.',
          'This page is part of that feedback loop. A reader can highlight a passage, attach a marginal note to the exact text, continue the reply thread, or open the page-level GitHub discussion. The annotation does not make the argument correct; it makes disagreement specific, durable, and reviewable.',
          'That is why human harnesses are becoming a first-class part of Polymetrics. The safest operator is not the one who promises never to make a mistake. It is the one whose tools expect mistakes and stop them before intention becomes mutation.',
        ],
      },
      {
        heading: 'The part where I ask for a star',
        body: [
          'The immediate work is certification. Connector inventory is broad; live, repeatable evidence is the next gate. We are also tightening approval scopes, improving bulk-write previews, extending dry-run diffs, and making destructive confirmation policy consistent across connector actions.',
          'Distribution is part of the same story: release binaries and a Homebrew path should make the first useful run short without bypassing provenance. The MCP surface should let agents discover the same typed operations without inventing a second, more permissive API.',
          'Shepherd is the next story. It starts with the question this post deliberately leaves open: what happens when the harness itself needs supervision? That deserves a full engineering note, so I am leaving it at the question rather than smuggling the implementation into this one.',
          'And now the shameless human bit: if this story made you laugh, wince, or remember a pull request whose scrollbar looked like a rounding error, please star the repository. A star will not certify 547 connectors or make make verify finish before lunch. It does tell me this strange, open-source plan is useful to someone outside my terminal.',
          'If you have survived your own giant-PR saga, leave a note on the paragraph that brought back the memories. I would genuinely like to hear the story, partly for research and partly so I know I am not the only person who has tried to review a small novel through GitHub.',
        ],
      },
    ],
  },
  {
    slug: 'one-cli-to-rule-them-all',
    title: 'One CLI To Rule Them All',
    description:
      'Why Polymetrics puts ETL, DuckDB SQL, reverse ETL, scheduling, and agent-safe JSON contracts behind one local binary.',
    publishedAt: '2026-07-02',
    updatedAt: '2026-07-02',
    readingTime: '5 min read',
    category: 'Product essay',
    tags: ['local-first ETL', 'DuckDB', 'reverse ETL', 'AI agents'],
    summary:
      'A single binary is easier to install, easier to audit, easier to automate, and easier for AI agents to operate without hidden infrastructure.',
    sections: [
      {
        heading: 'The data loop should not need a platform team',
        body: [
          'Most operational data work follows the same shape: extract from a source, land it somewhere queryable, decide what should happen, then write the result back to the systems where work happens.',
          'Polymetrics makes that loop a command-line workflow. The same binary owns connector setup, local storage, DuckDB queries, reverse ETL planning, approval, execution, and structured JSON output.',
        ],
      },
      {
        heading: 'Why a CLI is the right control plane',
        body: [
          'A CLI is portable, scriptable, observable, and easy to put under source control. It works on a laptop, in CI, in cron, in a container, and inside an AI agent loop without changing the interface.',
          'That matters because data workflows fail at the edges: credentials, schema drift, rate limits, approvals, retries, and audit trails. Keeping those edges in one binary keeps the mental model small.',
        ],
        points: [
          'No server to deploy before the first sync.',
          'No separate SQL service for local analysis.',
          'No second product just to write data back.',
          'No special agent API when the CLI already emits JSON.',
        ],
      },
      {
        heading: 'The loop in one place',
        body: [
          'Polymetrics treats extract, query, and act as one product surface. That lets a developer test the full workflow locally before promoting it to automation.',
        ],
        code: `pm etl run --connection github --stream issues --json
pm query run --sql "select * from issues where state = 'open'" --json
pm reverse plan sync --source-table stale_issues --destination github:write --json`,
      },
      {
        heading: 'What this unlocks',
        body: [
          'The product goal is simple: make serious data automation feel as lightweight as a Unix tool. The long-term moat is a broad connector catalog, a predictable local runtime, and safe write paths that both humans and agents can understand.',
        ],
      },
    ],
  },
  {
    slug: 'agent-native-data-workflows',
    title: 'Agent-Native Data Workflows Need Boring Contracts',
    description:
      'How JSON envelopes, stable exit codes, local credentials, and approval-gated writes make CLI automation safer for LLM agents.',
    publishedAt: '2026-07-02',
    updatedAt: '2026-07-02',
    readingTime: '6 min read',
    category: 'Engineering note',
    tags: ['AI agents', 'JSON', 'approval gates', 'CLI design'],
    summary:
      'Agentic data tools should be predictable before they are powerful. The contract must make success, failure, and mutation explicit.',
    sections: [
      {
        heading: 'Agents need less magic, not more',
        body: [
          'LLM agents are good at planning across tools, but they are fragile when the tools speak in prose, hide state, or mutate systems without a preview. A data CLI should expose a narrow, explicit contract.',
          'Polymetrics keeps the agent interface boring: JSON on stdout, logs on stderr, stable exit codes, and no destination writes until a plan is approved.',
        ],
      },
      {
        heading: 'The minimum safe contract',
        body: [
          'The agent should never need to parse progress bars or infer whether an operation succeeded from a sentence. Every command should make its status, resource identifiers, warnings, and retryability machine-readable.',
        ],
        points: [
          'Structured output with versioned envelopes.',
          'Exit codes that separate usage, validation, auth, connector, runtime, policy, and internal errors.',
          'Credentials referenced by name, never printed.',
          'Writes split into plan, preview, approve, and execute.',
        ],
      },
      {
        heading: 'Approval is a product feature',
        body: [
          'Reverse ETL is where automation becomes risky. The product should show the exact intended change before it touches the destination. That is useful for humans, and it is essential for agents.',
        ],
        code: `pm reverse plan sync --source-table candidates --destination github:write --json
pm reverse preview <plan-id> --json
pm reverse run <plan-id> --approve <token> --json`,
      },
      {
        heading: 'A better default for automation',
        body: [
          'The CLI becomes the shared contract for humans, CI, cron, and agent pods. There is no separate agent-only surface to secure, document, or debug.',
        ],
      },
    ],
  },
  {
    slug: 'local-first-data-engine',
    title: 'Local-First Data Pipelines Without Warehouse Sprawl',
    description:
      'A practical case for keeping extraction, local storage, analytical SQL, and write-back automation close to the developer.',
    publishedAt: '2026-07-02',
    updatedAt: '2026-07-02',
    readingTime: '5 min read',
    category: 'Architecture',
    tags: ['local-first', 'ETL', 'warehouse', 'open source'],
    summary:
      'Local-first does not mean toy. It means the first useful run happens on your machine, with production paths available when the workflow earns them.',
    sections: [
      {
        heading: 'Start local, promote deliberately',
        body: [
          'A surprising amount of data automation starts as a question: pull a few records, inspect a schema, join it with another source, then decide whether the result should become a scheduled workflow.',
          'If the first step requires a hosted warehouse, a worker deployment, and multiple service accounts, the experiment becomes heavier than the question.',
        ],
      },
      {
        heading: 'DuckDB changes the default',
        body: [
          'Embedded analytical SQL lets a CLI run real joins, aggregations, and window functions without asking the user to provision a database first. That makes local data work fast enough for exploration and structured enough for automation.',
        ],
        points: [
          'Extract connector data into a local warehouse.',
          'Run SQL near the files and credentials.',
          'Promote only the useful workflow to schedule or CI.',
          'Keep sensitive data on the developer machine unless a workflow intentionally writes it elsewhere.',
        ],
      },
      {
        heading: 'Local does not mean isolated',
        body: [
          'The same commands can run in cron, GitHub Actions, a container, or a managed runner later. Local-first is a product sequence, not a deployment ceiling.',
        ],
      },
      {
        heading: 'The open-source opportunity',
        body: [
          'Developers already understand the shape of command-line tools. A repo that explains its value clearly, documents the first run, exposes crawlable examples, and publishes useful technical essays gives search engines and AI answer engines real evidence to cite.',
        ],
      },
    ],
  },
];

export function getBlogPost(slug: string): BlogPost | undefined {
  return BLOG_POSTS.find((post) => post.slug === slug);
}

export function blogUrl(slug: string): string {
  return `/blog/${slug}`;
}
