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
WEBSITE_IMAGE="ghcr.io/karthik-sivadas/polymetrics-cli/website:<sha>" \
  website/deploy/deploy-podman-quadlet.sh
```
