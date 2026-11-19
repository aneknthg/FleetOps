# FleetOps Lab

FleetOps Lab is a platform operations lab: Terraform-managed Kubernetes, GitOps delivery with Argo CD, a Go operations CLI, vulnerability scanning, Prometheus/Grafana observability, and incident runbooks.

## Architecture

```text
GitHub repository (source of truth)
|-- terraform/         -> local kind clusters; optional AWS EKS module
|-- gitops/            -> Argo CD apps, dev auto-sync, prod manual sync
|-- tools/fleetctl/    -> Go inventory, audit, image listing, safe drain
|-- sample-app/        -> Go HTTP service and Prometheus metrics
`-- .github/workflows -> Terraform, Go, YAML, Kustomize, shell, Trivy checks

Feature branch -> pull request checks -> reviewed merge
                                      |-> Terraform changes are applied manually
                                      `-> Argo CD reconciles Git state to clusters
```

## What this demonstrates

| Platform work | FleetOps evidence |
|---|---|
| Infrastructure as code | Reusable kind module; optional VPC/EKS module with remote state setup |
| GitOps delivery | Root Application, ApplicationSet, Kustomize overlays; dev auto-sync and prod manual sync |
| Fleet operations | `fleetctl` context inventory, node/image reports, parallel audit, Eviction API drain |
| Security pipeline | Trivy image/config scans, Dependabot ecosystems, documented exception policy |
| Reliability | Prometheus alert rules, Grafana dashboard, runbooks, drill evidence templates |

## Quick start: local clusters

Install Docker, Terraform 1.10.5+, GNU Make, Go 1.27+, kind, kubectl, Helm, and Argo CD CLI. On Windows, use Ubuntu on WSL2 with Docker Desktop's WSL integration. The full workstation and credential checklist is in [docs/local-setup.md](docs/local-setup.md).

```sh
make up
kubectl --context kind-dev get nodes
make down
```

The clusters are named `mgmt`, `dev`, and `prod`; the pinned node image is `kindest/node:v1.36.6`. Follow [terraform/README.md](terraform/README.md) for individual environments and the separate, optional EKS workflow.

## GitOps setup

Build and publish the sample image to `ghcr.io/aneknthg/fleetops-sample-app:0.1.0` before deploying. The repository does not contain registry credentials. See [gitops/README.md](gitops/README.md) and [docs/local-setup.md](docs/local-setup.md) for the bootstrap flow. Argo CD uses the management cluster and internal kind endpoints to reach the workload clusters.

The sample app provides `/healthz`, `/version`, and `/metrics`. Prometheus scrapes the dev service; four alerts link to runbooks, and Grafana loads a dashboard from a labeled ConfigMap.

## `fleetctl`

Build and use the CLI from `tools/fleetctl/`:

```sh
cd tools/fleetctl
go mod tidy
go build -o fleetctl .
./fleetctl clusters
./fleetctl audit --all
```

See [tools/fleetctl/README.md](tools/fleetctl/README.md) for all five commands and output examples. Drain previews are the default; `--yes` opts into cordon and Eviction API requests that respect PDBs.

## CVE measurements

CI builds and scans the app image and scans Terraform and GitOps configuration. Dependabot checks Docker, Go modules, GitHub Actions, and Terraform weekly. The before/after table remains pending real Trivy scans; no counts are invented.

| Measurement | Critical | High | Fixable |
|---|---:|---:|---:|
| Baseline image | Pending scan | Pending scan | Pending scan |
| Updated image | Pending Dependabot/fix exercise | Pending Dependabot/fix exercise | Pending Dependabot/fix exercise |

Run [scripts/cve-report.sh](scripts/cve-report.sh) after clusters are available to create a fleet report.

## Reliability drills

The four planned drills and evidence status are listed in [docs/failure-drills.md](docs/failure-drills.md). GitOps proof exercises are tracked in [docs/gitops-proofs.md](docs/gitops-proofs.md). Copy [docs/postmortems/TEMPLATE.md](docs/postmortems/TEMPLATE.md) after running each drill and record actual timestamps and outcomes. Live drills have not been run in this environment.

Runbooks: [bad deployment rollback](docs/runbooks/rollback-bad-deploy.md), [node not ready](docs/runbooks/node-not-ready.md), [node image rotation](docs/runbooks/node-image-rotation.md), and [CVE triage](docs/runbooks/cve-triage.md).

## Decisions and limitations

See [docs/decisions/](docs/decisions/) for why the lab uses kind, Argo CD, and safe node draining. The initial implementation is merged into `main` as [PR #1](https://github.com/aneknthg/FleetOps/pull/1). kind nodes share the host kernel; three local clusters do not demonstrate 1,000-cluster scale. AWS EKS is coded but remains unprovisioned. Branch protection is not enabled yet and should be configured by a repository administrator. `.github/workflows/terraform-plan.yml` is manual and plan-only for AWS EKS.

## Cost and safety

Local kind uses laptop resources. EKS resources incur AWS charges, including its control plane and NAT Gateway. Create a $10 monthly budget alert and review a Terraform plan before any cloud apply. Destroy the EKS resources the same day and check the AWS account for leftover resources.

## AI-assisted development

Codex helped draft the initial Terraform, GitOps, Go CLI, security workflow, and runbook scaffolding. One real debugging example: GitHub Actions caught an audit version-regex capture-count bug and a slice-flattening compile error; the implementation was corrected from the failing logs. Review and understand the code before presenting it as personal operational experience.
