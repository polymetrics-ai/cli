# Website Deployment

The website pipelines build and test the Next.js app on pull requests and pushes that touch `website/`. Production deployment is limited to `main` and runs on the Polymetrics origin VPS through a Tailscale-connected self-hosted GitHub runner.

## Flow

- GitHub Actions runs `gen:website-data`, generated-file drift checks, `typecheck`, `build`, and a Docker image build. Pushes to `main` publish `ghcr.io/<owner>/<repo>/website:<sha>` and `:main`.
- GitLab CI runs the same website checks, builds the Docker image, and pushes `$CI_REGISTRY_IMAGE/website:<sha>` plus `:main` only on `main`.
- Deployment runs `website/deploy/deploy-podman-quadlet.sh` on the origin runner. The script pulls the CI-built GHCR tag, resolves it to an immutable digest, updates the rootless Quadlet `Image=...@sha256:...`, restarts `cli-polymetrics.service`, checks loopback health, and verifies the public Cloudflare Tunnel URL.
- `website/deploy/deploy-image.sh` and the Kubernetes manifests are retained for future Kubernetes environments, but they are not the active Polymetrics origin path.

## Runner Assumptions

GitHub deploy runner:

- Labels: `self-hosted`, `linux`, `tailscale`, `polymetrics-website`.
- Runs on the origin VPS as the `deploy` user.
- Joined to the tailnet.
- Has rootless Podman, `curl`, and user systemd available.
- Has linger enabled for `deploy` so `systemctl --user` can restart the Quadlet from CI.

GitLab deploy runner:

- Tags: `tailscale`, `polymetrics-website`.
- Joined to the tailnet and able to reach the target Kubernetes API.
- Can run the `bitnami/kubectl` image, or uses a shell executor with `kubectl` already installed.

GitLab image builds use Docker-in-Docker. The runner used by `website:image` must allow privileged Docker service jobs, or the job should be swapped to the local registry build mechanism used by your GitLab installation.

## GitHub Variables And Secrets

Required:

- `GITHUB_TOKEN`: provided by GitHub Actions; used to push to GHCR from `main`.

Optional:

- Variable `WEBSITE_SERVICE`: defaults to `cli-polymetrics`.
- Variable `WEBSITE_QUADLET`: defaults to `/home/deploy/.config/containers/systemd/cli-polymetrics.container`.
- Variable `WEBSITE_HEALTH_URL`: defaults to `http://127.0.0.1:18081/`.
- Variable `WEBSITE_PUBLIC_URL`: defaults to `https://cli.polymetrics.ai/`. Set to an empty value to skip the public URL check.
- Variable `WEBSITE_ROLLOUT_TIMEOUT`: defaults to `120s`.

## Database Sidecar And App Secrets

Blog comments/bookmarks and sign-in use a Postgres sidecar next to the website
container. One-time setup on the origin VPS (as the `deploy` user):

1. Install the shared network and database Quadlets from this directory:

   ```sh
   cp cli-polymetrics.network cli-polymetrics-db.container ~/.config/containers/systemd/
   ```

2. Create the env files (mode `0600`) under `~/.config/cli-polymetrics/`:

   `db.env`:

   ```sh
   POSTGRES_DB=website
   POSTGRES_USER=website
   POSTGRES_PASSWORD=<generated>
   ```

   `website.env`:

   ```sh
   DATABASE_URL=postgres://website:<generated>@cli-polymetrics-db:5432/website
   BETTER_AUTH_SECRET=<openssl rand -hex 32>
   BETTER_AUTH_URL=https://cli.polymetrics.ai
   GITHUB_CLIENT_ID=...
   GITHUB_CLIENT_SECRET=...
   GOOGLE_CLIENT_ID=...
   GOOGLE_CLIENT_SECRET=...
   LINKEDIN_CLIENT_ID=...
   LINKEDIN_CLIENT_SECRET=...
   ADMIN_EMAILS=you@example.com
   ```

   Social providers are optional — sign-in offers whichever providers have
   credentials configured.

3. Add these lines to the website Quadlet
   (`~/.config/containers/systemd/cli-polymetrics.container`). The deploy
   script only rewrites `Image=`, so they survive every deploy and rollback:

   ```ini
   Network=cli-polymetrics.network
   EnvironmentFile=%h/.config/cli-polymetrics/website.env
   ```

4. Reload and start:

   ```sh
   systemctl --user daemon-reload
   systemctl --user start cli-polymetrics-db
   systemctl --user restart cli-polymetrics
   ```

Schema migrations apply automatically on the website's first database request
(advisory-locked, idempotent) — no deploy-pipeline step is needed.

OAuth callback URLs to register on each provider:

- Production: `https://cli.polymetrics.ai/api/auth/callback/{github,google,linkedin}`
- Local dev: `http://localhost:3000/api/auth/callback/{github,google,linkedin}`

Backups: `podman exec cli-polymetrics-db pg_dump -U website website > backup.sql`
(cron it; the `cli-polymetrics-pgdata` volume holds the live data).

## GitLab Variables

Required built-ins for GitLab Container Registry:

- `CI_REGISTRY`
- `CI_REGISTRY_IMAGE`
- `CI_REGISTRY_USER`
- `CI_REGISTRY_PASSWORD`

Optional project or environment variables:

- `WEBSITE_KUBECONFIG_B64`: masked and protected base64-encoded kubeconfig, only needed when the deploy runner does not already have kubeconfig.
- `WEBSITE_KUBE_CONTEXT`
- `WEBSITE_NAMESPACE`: keep it aligned with the namespace in these manifests.
- `WEBSITE_DEPLOYMENT`
- `WEBSITE_CONTAINER`
- `WEBSITE_IMAGE_PULL_SECRET`
- `WEBSITE_ROLLOUT_TIMEOUT`

## Private Registry Pulls

For GHCR pulls, the deploy job logs Podman into `ghcr.io` using the job-scoped `GITHUB_TOKEN` with `packages: read`. No long-lived registry password is stored on the VPS.

## Manual Deploy

Use the same script from the origin VPS as `deploy`:

```bash
WEBSITE_IMAGE="ghcr.io/polymetrics-ai/cli/website:<sha>" \
  website/deploy/deploy-podman-quadlet.sh
```
