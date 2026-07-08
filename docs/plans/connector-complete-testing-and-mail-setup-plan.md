# Connector Complete Testing & Mail Setup Plan

## Part 1: Best approach for testing every stream and every PM function

### The problem

The existing `pm connectors certify` Runner tests **one stream** (the first or
`--stream`-specified) and **one write pairing** (the first `create_X`/`delete_X`
pair). It does NOT test:
- all 37+ read streams individually
- all 231 write actions (each create/delete/update/close pair)
- binary downloads (release assets, tarballs, logs)
- direct reads (single-record lookups)
- flow + schedule roundtrips per stream
- the reverse-ETL plan→preview→approve→execute for gated writes

So "certify passed" means "one stream + one write worked" — not "every
function works." We need a **comprehensive sweep** that validates every
stream, every write action, every binary, every direct read — and handles the
**create-then-delete lifecycle** (you can't delete what doesn't exist; you
can't verify a create without reading it back; binary needs a file to
download).

### Deep analysis: script vs pi agent

| Dimension | Bash/Python script | pi agent (TS, LLM-driven) |
|---|---|---|
| **Determinism** | ✅ fully deterministic, same input → same output | ⚠️ LLM may choose different paths; needs strict skill constraints |
| **Speed** | ✅ minutes (no LLM calls) | ❌ slower (each decision = an LLM call) |
| **Cost** | ✅ free | ❌ token cost per run |
| **Handles unexpected API errors** | ❌ brittle (hardcoded expectations) | ✅ can reason: "404? maybe the resource name changed; retry with a different name" |
| **Handles create→delete ordering** | ✅ if hardcoded correctly | ✅ can infer: "create_label needs a name; delete needs the same name; if create failed, skip delete" |
| **Handles schema mismatches** | ❌ fails opaquely | ✅ can read the error, inspect the schema, adjust the record |
| **Handles gated writes (approval)** | ✅ if the gate is scripted | ✅ can reason: "this is destructive; I need --confirm" |
| **Auditability** | ✅ script is the spec | ⚠️ the skill is the spec; the agent's execution is logged |
| **Reproducibility for CI** | ✅ perfect | ⚠️ needs a "deterministic mode" |

### The hybrid answer: **scripted engine + pi agent for anomaly resolution**

The best approach is **not** "script OR agent" — it's a **scripted test engine
that generates a per-connector test matrix from the connector's own metadata**
(streams.json, writes.json, operations.json, cli_surface.json), runs every test
deterministically, and **escalates failures to a pi agent** for diagnosis +
adaptive retry. This gives:

1. **Determinism + speed + cost** (the script runs 37 streams + 231 writes in
   minutes, no LLM calls).
2. **Adaptive failure handling** (the pi agent diagnoses unexpected errors,
   inspects the schema, adjusts records, retries — only on failure).
3. **The create→delete lifecycle** is handled by the engine's pairing table
   (already exists: `InferPairing` derives `create_X`/`delete_X`/`close_X`/
   `archive_X` pairs automatically from the write action names).

### The engine: `pm connectors certify --full` (extends the existing Runner)

```
pm connectors certify github --full \
  --connection github-dev --secret-env token=PM_GITHUB_DEV_TOKEN \
  --repo polymetrics-ai/<dev>-parity \
  --json --keep-work
```

`--full` triggers **per-stream + per-write-action + per-binary + per-direct-read**
testing, not just the first of each. The engine:

#### 1. Read sweep (every stream)

For each stream in `streams.json`:
```
pm github <stream> --connection github-dev --stream <name> --limit 5 --json
```
Records: pass/fail, record count, schema conformance (every record validates
against the stream's schema), pagination (does `--limit` work?), incremental
cursor (does `--since <cursor>` work?). If a stream returns 0 records, it's
flagged "untested" (not failed) — some streams need seed data.

#### 2. Write sweep (every create→delete pair)

For each write action in `writes.json`, derive the pairing:
- `create_X` → `delete_X` / `close_X` / `archive_X` (via `InferPairing`).
- `update_X` → needs a pre-existing entity (create first, update, then delete).
- `close_X` / `reopen_X` → create first, close, reopen, then delete.

For each pair, the lifecycle:
```
# 1. Create: generate a record from the write action's record_schema
pm github reverse plan --connection github-dev --stream create_X --record '<generated>' --json
# 2. Preview: redacted preview
pm github reverse preview <plan-id> --json
# 3. Execute: --approve (and --confirm for gated writes, once slice 2 lands)
pm github reverse run <plan-id> --approve $TOKEN --json
# 4. Verify: read back via the verify stream (e.g. issues, labels)
pm github <verify-stream> --connection github-dev --stream <name> --json
# 5. Cleanup: delete/close/archive the created entity
pm github reverse plan --connection github-dev --stream delete_X --record '<id>' --json
pm github reverse run <plan-id> --approve $TOKEN --json
# 6. Cleanup-verify: read-back confirms the entity is gone
```

If create fails → skip delete (nothing to clean). If delete fails → record a
**leak** (the sweeper catches it later). Every created entity gets a tag
`pm-cert-<connector>-<runid>-<ts>` so the sweeper can find and clean orphans.

#### 3. Binary sweep (every binary_download operation)

For each `binary_download` operation:
```
pm github <cmd> --connection github-dev --operation <op-id> --output /tmp/cert-bin --json
```
Records: pass/fail, bytes downloaded, MIME type, `max_bytes` enforced. Binary
tests need a **pre-existing resource** (a release with assets, a workflow run
with logs) — the engine creates a seed (e.g. a draft release with a dummy
asset) before the download, then cleans it up.

#### 4. Direct-read sweep (every direct_read CLI command)

For each `direct_read` CLI command:
```
pm github <cmd> --connection github-dev --operation <op-id> --json
```
Records: pass/fail, JSON shape. Direct reads that need a path param (e.g.
`issue view --issue-number 1`) use the first record from the corresponding
read stream as input.

#### 5. Flow + schedule roundtrip (per stream)

The existing `stageFlowRoundtrip` + `stageScheduleRoundtrip` — extended to run
for **each stream**, not just the capture stream.

#### 6. Pi-agent escalation (on failure)

If any test fails, the engine writes the failure to a JSON file and invokes a
**pi agent** in diagnostic mode:
```
pi --tools read,bash,grep,find,ls --agentScope both --confirmProjectAgents false \
  -p "Diagnose this certify failure: <failure-json>. Inspect the connector
      defs, the error, the schema. Propose a fix or an adaptive retry. Do not
      execute live writes."
```
The pi agent reads the connector defs + the error + the schema, reasons about
the mismatch (e.g. "the record_schema requires `owner` but the CLI flag is
`--repo`"), and either proposes a fix (a PR) or an adaptive retry (adjusted
record). This is where the pi agent's **knowledge of create/delete/binary
lifecycle** adds value — it understands "I created an issue, so I need its
number to delete it; the number is in the create response."

### Why the pi agent is NOT the primary runner

- **Cost**: 37 streams × 231 writes × LLM calls = thousands of tokens per run.
- **Speed**: minutes (script) vs hours (agent).
- **Determinism**: CI needs the same result every time; an LLM may flake.
- **The lifecycle is already declarative**: `InferPairing` + `record_schema`
  data generation + the verify stream handle create→delete without reasoning.

The pi agent is the **diagnostic layer**, not the execution layer.

### Output: the certificate

```json
{
  "connector": "github",
  "verifiedBy": "polymetrics-ai",
  "streams": { "tested": 37, "passed": 37, "untested": 0 },
  "writes": { "tested": 231, "passed": 228, "failed": 3, "leaks": 0 },
  "binary": { "tested": 10, "passed": 10 },
  "direct_reads": { "tested": 164, "passed": 164 },
  "flow": { "tested": 37, "passed": 37 },
  "schedule": { "tested": 37, "passed": 37 },
  "overall": { "passed": true, "coverage": "440/442" }
}
```

Only `overall.passed == true && leaks == 0` earns the "verified by
organization" badge.

---

## Part 2: Stalwart mail server setup on the VPS (detailed)

### Prerequisites

- A VPS (the Polymetrics VPS) with a public IP and Tailscale installed.
- DNS control over `polymetrics.ai` (to set MX/SPF/DKIM/DMARC records).
- Root/sudo on the VPS.
- A Tailscale tail IP for the VPS (e.g. `100.x.y.z`).

### Step 1: DNS records (set BEFORE starting Stalwart)

In the `polymetrics.ai` DNS zone:

```
# MX record: mail.polymetrics.ai handles email
@                MX     10 mail.polymetrics.ai.

# A record for the mail host (public IP for public SMTP, or Tailscale IP for private-only)
mail             A      <VPS-PUBLIC-IP>

# SPF: only mail.polymetrics.ai is allowed to send
@                TXT    "v=spf1 mx -all"

# DMARC: quarantine failures, send reports to postmaster
_dmarc           TXT    "v=DMARC1; p=quarantine; adkim=s; aspf=s; rua=mailto:postmaster@polymetrics.ai"

# DKIM: Stalwart generates this on first run; publish the selector it prints
# (e.g. <selector>._domainkey  TXT  "v=DKIM1; k=rsa; p=<base64>")
```

If you want `github@polymetrics.ai` to be **reachable only via Tailscale** (not
publicly), set the MX to the Tailscale IP and skip the public A record — but
then GitHub can't send email to it (GitHub's servers can't reach a Tailscale
IP). **Recommendation: use the public IP for MX** so GitHub notifications
reach the mailbox; use Tailscale for the Mailspring IMAP/SMTP connection.

### Step 2: Install Stalwart on the VPS

```bash
# SSH to the VPS
ssh polymetrics@<vps>

# Download Stalwart (Rust, single binary)
curl -fsSL https://github.com/stalwartlabs/stalwart/releases/latest/download/stalwart-linux-x86_64.tar.gz \
  | sudo tar -xz -C /opt/stalwart

# Initialize (generates config + DKIM keys + admin password)
sudo /opt/stalwart/bin/stalwart -c /opt/stalwart/etc/config.toml --init

# Note the admin password it prints; change it after first login.
```

### Step 3: Configure Stalwart

Edit `/opt/stalwart/etc/config.toml`:

```toml
[server.listen.smtp]
bind = ["[::]:25", "[::]:465", "[::]:587"]

[server.listen.imap]
bind = ["[::]:993"]

[server.listen.https]
bind = ["[::]:443"]

[storage.data]
path = "/opt/stalwart/data"

[certificate.default]
cert = "/opt/stalwart/etc/certs/mail.polymetrics.ai.fullchain.pem"
key = "/opt/stalwart/etc/certs/mail.polymetrics.ai.privkey.pem"

# Or use Let's Encrypt (Stalwart can auto-provision):
[certificate.letsencrypt]
domain = "mail.polymetrics.ai"
```

Get a TLS certificate (if not using Stalwart's built-in Let's Encrypt):
```bash
sudo certbot certonly --standalone -d mail.polymetrics.ai
sudo cp /etc/letsencrypt/live/mail.polymetrics.ai/fullchain.pem /opt/stalwart/etc/certs/
sudo cp /etc/letsencrypt/live/mail.polymetrics.ai/privkey.pem /opt/stalwart/etc/certs/
```

### Step 4: Create the `github@polymetrics.ai` mailbox + app password

```bash
# Create the account
sudo /opt/stalwart/bin/stalwart -c /opt/stalwart/etc/config.toml account create \
  --name "Polymetrics GitHub" \
  --address github@polymetrics.ai \
  --type individual

# Create a dedicated APP PASSWORD for Mailspring (NOT the GitHub account password)
sudo /opt/stalwart/bin/stalwart -c /opt/stalwart/etc/config.toml credential add \
  --address github@polymetrics.ai \
  --name mailspring
# Note the app password it prints.
```

### Step 5: Start Stalwart

```bash
# As a systemd service
sudo systemctl enable stalwart
sudo systemctl start stalwart
sudo systemctl status stalwart

# Verify it's listening
sudo ss -tlnp | grep -E ':(25|465|587|993|443)\s'
```

### Step 6: Test email reception

```bash
# From another machine, send a test email to github@polymetrics.ai
echo "test" | mail -s "Stalwart test" github@polymetrics.ai

# Or use swaks:
swaks --to github@polymetrics.ai --from test@example.com --server mail.polymetrics.ai

# Check Stalwart logs
sudo journalctl -u stalwart -f
```

### Step 7: Configure Mailspring locally (you do this)

1. Open Mailspring → Add Account → IMAP.
2. **IMAP**: `mail.polymetrics.ai:993` (SSL), user `github@polymetrics.ai`, password = the **app password** from Step 4 (NOT the GitHub account password).
3. **SMTP**: `mail.polymetrics.ai:465` (SSL), same credentials.
4. Test: send + receive a test email.

### Step 8: Create the GitHub account (you do this)

1. Go to github.com → Sign up → use `github@polymetrics.ai` as the email.
2. Verify the email (check Mailspring for the GitHub verification email).
3. Add the account as a **collaborator** on `polymetrics-ai/cli` and on the dev repo `polymetrics-ai/<dev>-parity`.
4. Create a **fine-grained PAT** scoped to the dev repo: `issues: write, pull_requests: write, metadata: read, contents: read` (no admin/secrets/workflows).
5. Export the token: `export PM_GITHUB_DEV_TOKEN="<the PAT>"` and give it to me.

### Step 9: Register the credential in PM

```bash
pm connections create github-dev --connector github \
  --config owner=polymetrics-ai --config repo=<dev>-parity
pm credentials add github-dev-pat --connection github-dev \
  --from-env token=PM_GITHUB_DEV_TOKEN
pm connections test github-dev --json
```

### Step 10: Run the full certification

```bash
pm connectors certify github --full \
  --connection github-dev --secret-env token=PM_GITHUB_DEV_TOKEN \
  --repo polymetrics-ai/<dev>-parity \
  --json --keep-work
```

### Security notes

- The **app password** (Mailspring) is NOT the GitHub account password.
- The **GitHub PAT** is scoped to the dev repo only, never `cli` content writes.
- `pm` never stores the token value; it reads from `--from-env`.
- Stalwart logs don't contain the GitHub PAT.
- The mailbox is for GitHub notifications + account recovery only; `pm` uses the PAT, not the mailbox.

### Hardening

- Enable `fail2ban` on SMTP/IMAP ports.
- Set DMARC `p=reject` after verifying deliverability.
- Rotate the app password + PAT quarterly.
- Monitor Stalwart logs for brute-force attempts.
- Backup the Stalwart data directory (`/opt/stalwart/data`).
