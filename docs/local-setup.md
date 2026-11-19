# Local setup and credentials

This guide is written for a Windows workstation. Run the project commands inside Ubuntu on WSL2; the Terraform kind provider, Makefile, and bootstrap script expect a Linux/Bash environment.

## Local path: no AWS account or cloud credentials

You need:
- A GitHub account with read access to `aneknthg/FleetOps` (the owner account already has this).
- Windows 11 with WSL2 and Ubuntu 24.04, plus Docker Desktop with WSL integration enabled for that Ubuntu distribution.
- Terraform 1.10.5 or newer, GNU Make, Go 1.27 or newer, kind, kubectl, Helm, and the Argo CD CLI installed in Ubuntu/WSL.
- A GitHub personal access token with `write:packages` only if you will publish the sample image to GHCR. Argo CD's workload clusters need to pull that image; make the package public or configure image-pull credentials.

No AWS access key, Twilio account, or other service credential is used by the local kind path. Do not put tokens, kubeconfig files, Terraform state, or `.tfvars` files in Git.

### 1. Prepare WSL2 and Docker Desktop

From an elevated PowerShell terminal, install Ubuntu if WSL2 is not already set up:

```powershell
wsl --install -d Ubuntu-24.04
```

Install Docker Desktop for Windows, enable its WSL2 engine, and turn on WSL integration for Ubuntu 24.04. Open Ubuntu and verify `docker version` works without `sudo`. Install Git, GNU Make, and the project CLIs using the vendors' official installation instructions:

- [Terraform](https://developer.hashicorp.com/terraform/install)
- [Go](https://go.dev/doc/install)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)
- [Argo CD CLI](https://argo-cd.readthedocs.io/en/stable/cli_installation/)

Check versions and Docker connectivity before continuing:

```sh
docker version
terraform version
go version
kind version
kubectl version --client
helm version
argocd version --client
make --version
```

### 2. Clone, test, and publish the sample image

Run these commands in Ubuntu/WSL from your home directory:

```sh
git clone https://github.com/aneknthg/FleetOps.git
cd FleetOps
make test
docker build -t ghcr.io/aneknthg/fleetops-sample-app:0.1.0 ./sample-app
```

To push to GitHub Container Registry, create a token with package write permission and enter it at the prompt. The token is passed through stdin and is not stored in shell history:

```sh
read -rsp 'GHCR token: ' GHCR_TOKEN; printf '\n'
printf '%s' "$GHCR_TOKEN" | docker login ghcr.io --username aneknthg --password-stdin
unset GHCR_TOKEN
docker push ghcr.io/aneknthg/fleetops-sample-app:0.1.0
```

After the package is published, open its package settings on GitHub and set visibility to public for this lab, or configure Kubernetes image-pull credentials before deploying.

### 3. Create the local clusters and start GitOps

```sh
make up
kubectl --context kind-dev get nodes
./scripts/bootstrap.sh
kubectl --context kind-mgmt -n argocd get applications
kubectl --context kind-dev get pods -A
```

`make up` creates `mgmt`, `dev`, and `prod` kind clusters. The bootstrap script installs Argo CD on `mgmt`, registers `dev` and `prod`, then applies the root Application. Argo CD deploys the sample application and dev monitoring from Git.

Remove the clusters when finished to release laptop resources:

```sh
make down
```

### 4. Optional AWS EKS path

Do not create AWS resources until you have:
- An AWS account and a verified sign-in path. Prefer AWS IAM Identity Center / `aws configure sso`; avoid long-lived access keys where possible.
- An active AWS Budgets alert at $10/month.
- An AWS identity with permissions to create the bootstrap S3 state bucket and the EKS/VPC/EC2/IAM resources in the Terraform plan.
- A globally unique S3 bucket name for Terraform state.
- An AWS IAM role trusted for GitHub Actions OIDC if you use `.github/workflows/terraform-plan.yml`; do not store AWS keys in GitHub secrets.
- Your current public IPv4 address in CIDR form, such as `203.0.113.5/32`, to restrict the EKS API endpoint.

Verify the selected identity before following [terraform/README.md](../terraform/README.md):

```sh
aws sts get-caller-identity
```

The workflow is plan-only. Terraform apply, cluster upgrades, and destroy are manual exercises. EKS, NAT Gateway, workers, public IPv4, and data transfer can incur charges. Review every plan and destroy the resources the same day after an exercise.

## What has and has not been run

GitHub Actions has validated Terraform, Go, YAML/Kustomize, shell, builds, and Trivy scans. The local kind clusters, Argo CD bootstrap, GHCR publication, AWS EKS, monitoring UI, and reliability drills still require your workstation/account and have not been executed by this project setup.
