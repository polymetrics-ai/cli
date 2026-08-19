# Discussion log — issue #4015 GitHub declared-command parity

## Invocation

`scripts/gsd prompt discuss-phase issue-4015-github-cli-parity-50 --auto`

The repo-local adapter resolved the official command. The workflow ran inline because the
repository's canonical single-worker contract forbids spawning lifecycle roles. The launch brief
explicitly requested autonomous execution, so `--auto` selected the repository-safe defaults.

## Decisions resolved from the brief and repository canon

1. **Breadth:** inspect and verdict all 50 commands; implement every command admitted by a typed
   existing connector or local-workflow capability.
2. **Provider evidence:** verify current GitHub documentation for every API-backed claim. Empty
   surface metadata never proves absence.
3. **Runtime shape:** use fixed REST/GraphQL operations and existing declarative patterns. Shared
   foundation gaps are not hidden in GitHub-specific code.
4. **Local workflows:** dependency-free does not authorize arbitrary subprocess execution. Local
   commands without a typed workflow stay declared and receive exact product-boundary evidence.
5. **Raw API:** the upstream `gh api` command conflicts with the binding prohibition on generic
   HTTP write tools. It remains declared but not implemented unless a fixed typed operation can
   replace the unrestricted surface (which would no longer be `gh api` parity).
6. **Secrets:** `auth token` cannot be implemented because its intended behavior is to print a
   credential, which the binding agent/product rules prohibit.
7. **Writes and cleanup:** live mutation proof is authorized only in the named certification org
   and repository, with disposable prefixes and independent post-delete absence assertions.
8. **Documentation:** runtime help, generated manual, connector docs, website references, and
   discovery metadata are updated or marked not applicable for each implementation family.

## Deferred decisions

No product decision is deferred. A provider endpoint with real-money, real-person, public
visibility, third-party repository, or missing-scope implications remains unexecuted and receives a
truthful live-proof limitation unless the bounded operation can be proven safely within the launch
brief's authority.

