# FleetOps Lab

FleetOps Lab is a hands-on platform operations project: Terraform-managed Kubernetes, GitOps delivery with Argo CD, a Go operations CLI, image security checks, and operational runbooks.

## Status

The repository is being built in stages. The initial scaffold and reusable Terraform module for three local kind clusters are in place. Live Kubernetes and AWS exercises require the relevant local tools, accounts, and credentials.

## Architecture

```text
Git change -> pull request checks -> reviewed merge
                                  |-> Terraform provisions clusters
                                  `-> Argo CD reconciles gitops/ to clusters
fleetctl inspects and drains nodes; Trivy scans images and configuration.
```

## Quick start

Install Docker Engine or Docker Desktop, Terraform 1.6+, and GNU Make. Run `make up` to create the `mgmt`, `dev`, and `prod` clusters; run `make down` to destroy them. Review Terraform plans before production or cloud changes. See [terraform/README.md](terraform/README.md) for direct commands.

## Repository layout

- `terraform/` — reusable local kind and AWS EKS infrastructure modules
- `gitops/` — Argo CD bootstrap and Kubernetes application definitions
- `tools/fleetctl/` — Go operations CLI
- `sample-app/` — small instrumented Go service
- `scripts/` — setup and operational scripts
- `docs/` — runbooks, postmortems, and architecture decisions
- `.github/workflows/` — CI and security automation

## Project notes

The local kind environment is a lab, not a production fleet. kind nodes share the host kernel, and the initial exercises cover a small number of clusters. AWS resources are optional and must be cost controlled.

