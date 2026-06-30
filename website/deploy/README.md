# Website Deployment

The website pipelines build and test the Next.js app on pull requests and pushes that touch `website/`. Deployment is limited to `main` and runs on self-hosted infrastructure that can reach the Kubernetes API over Tailscale.

## Flow

- GitHub Actions runs `gen:website-data`, generated-file drift checks, `typecheck`, `build`, and a Docker image build. Pushes to `main` publish `ghcr.io/<owner>/<repo>/website:<sha>` and `:main`.
- GitLab CI runs the same website checks, builds the Docker image, and pushes `$CI_REGISTRY_IMAGE/website:<sha>` plus `:main` only on `main`.
- Deployment runs `website/deploy/deploy-image.sh`, applies the namespace/deployment/service manifests, sets the CI-built image, changes `imagePullPolicy` to `IfNotPresent`, optionally sets an image pull secret name, and waits for rollout.

## Runner Assumptions

GitHub deploy runner:

- Labels: `self-hosted`, `linux`, `tailscale`, `polymetrics-website`.
- Joined to the tailnet and able to reach the target Kubernetes API.
- Has `kubectl` installed.
- Has a working kubeconfig already on disk, or receives one through `WEBSITE_KUBECONFIG_B64`.

GitLab deploy runner:

- Tags: `tailscale`, `polymetrics-website`.
- Joined to the tailnet and able to reach the target Kubernetes API.
- Can run the `bitnami/kubectl` image, or uses a shell executor with `kubectl` already installed.

GitLab image builds use Docker-in-Docker. The runner used by `website:image` must allow privileged Docker service jobs, or the job should be swapped to the local registry build mechanism used by your GitLab installation.

## GitHub Variables And Secrets

Required:

- `GITHUB_TOKEN`: provided by GitHub Actions; used to push to GHCR from `main`.
- Variable `WEBSITE_DEPLOY_ENABLED`: set to `true` after the Tailscale self-hosted runner is registered. Until then, the deploy job stays skipped instead of queuing forever.

Optional:

- Secret `WEBSITE_KUBECONFIG_B64`: base64-encoded kubeconfig, only needed when the deploy runner does not already have kubeconfig.
- Variable `WEBSITE_KUBE_CONTEXT`: kube context to select before deployment.
- Variable `WEBSITE_NAMESPACE`: defaults to `polymetrics-website`; keep it aligned with the namespace in these manifests.
- Variable `WEBSITE_DEPLOYMENT`: defaults to `polymetrics-website`.
- Variable `WEBSITE_CONTAINER`: defaults to `website`.
- Variable `WEBSITE_IMAGE_PULL_SECRET`: Kubernetes secret name to attach to the deployment for private registry pulls.
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

For private GHCR or GitLab registry images, create a Kubernetes image pull secret out of band and set `WEBSITE_IMAGE_PULL_SECRET` to that secret name. Keep tokens in environment variables or secret stores.

Example shape:

```bash
kubectl -n polymetrics-website create secret docker-registry website-registry \
  --docker-server="$REGISTRY_HOST" \
  --docker-username="$REGISTRY_USERNAME" \
  --docker-password="$REGISTRY_TOKEN"
```

## Manual Deploy

Use the same script from a Tailscale-connected machine with `kubectl` configured:

```bash
WEBSITE_IMAGE="$REGISTRY_HOST/$REGISTRY_PATH/website:$IMAGE_TAG" \
  website/deploy/deploy-image.sh
```
