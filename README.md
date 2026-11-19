# FleetOps Lab

FleetOps Lab is a hands-on platform operations project: Terraform-managed Kubernetes, GitOps delivery with Argo CD, a Go operations CLI, image security checks, and operational runbooks.

## Status

The repository is being built in stages. Local kind infrastructure, GitOps scaffolding, the sample Go service, and `fleetctl` are implemented on [PR #1](https://github.com/aneknthg/FleetOps/pull/1). Live cluster provisioning and AWS exercises have not been run in this environment.

## Architecture

```text
Git change -> pull request checks -> reviewed merge
                                  |-> Terraform provisions clusters
                                  `-> Argo CD reconciles gitops/ to clusters
fleetctl inventories, audits, and drains; Trivy scans images and configuration.
```

## Quick start

Install Docker Engine or Docker Desktop, Terraform 1.6+, GNU Make, Go 1.26+, kind, kubectl, and Helm. Run `make up` to create the `mgmt`, `dev`, and `prod` clusters; run `make down` to destroy them. The workflow is documented in [terraform/README.md](terraform/README.md).

## `fleetctl`

Build with `cd tools/fleetctl && go build -o fleetctl .`. The command reference and example outputs are in [tools/fleetctl/README.md](tools/fleetctl/README.md). The CLI provides `clusters`, `nodes`, `images`, `audit`, and dry-run-by-default `drain` commands.

## CVE reporting

The CI workflow scans the sample image and Terraform/Kubernetes configuration for fixable HIGH and CRITICAL findings. To generate a fleet report, set `FLEETCTL_CONTEXT` and run `scripts/cve-report.sh` with `fleetctl`, `trivy`, and `jq` installed.

| Measurement | Critical | High | Fixable |
|---|---:|---:|---:|
| Baseline image scan | Pending local Trivy run | Pending local Trivy run | Pending local Trivy run |
| Patched image scan | Pending Dependabot/update exercise | Pending Dependabot/update exercise | Pending Dependabot/update exercise |

No vulnerability counts are claimed until the image is built and scanned.

## Failure drills

Run the drills only after the local clusters and monitoring stack are available. Record real timelines and outcomes in `docs/postmortems/`; no postmortem is fabricated before a drill is performed.

## Decisions and limitations

- kind shares the host kernel; kernel findings are represented through node inventory and the configured minimum kernel check.
- The lab begins with three local clusters. It does not demonstrate an actual 1,000-cluster fleet.
- EKS resources are optional and should only be applied with an AWS budget alert and an explicit destroy plan.

## Cost

The local kind workflow uses local compute. EKS, NAT gateways, and worker nodes incur AWS charges; the EKS exercise remains unprovisioned until the account budget and credentials are ready.

## AI-assisted development

Codex was used to draft the repository scaffold, Terraform module, GitOps manifests, and `fleetctl` implementation. Review the code and explain each operational choice before using it as interview evidence.
