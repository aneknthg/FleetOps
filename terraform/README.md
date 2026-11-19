# Terraform local clusters

The `kind-cluster` module creates one control-plane node and a configurable number of workers. The local environments keep independent local Terraform state so each cluster can be planned and destroyed independently.

## Prerequisites

Install Docker Engine or Docker Desktop, kind, Terraform 1.6+, and GNU Make. The Terraform kind provider creates kind clusters through the local container runtime. This environment does not currently have those tools installed, so provisioning has not been run here.

## Create and remove all local clusters

```sh
make up
kubectl --context kind-dev get nodes
make down
```

The default topology creates `mgmt`, `dev`, and `prod` with one control plane and two workers each. The node image is pinned to `kindest/node:v1.36.6`. Changing a kind node image requires cluster recreation; do not treat this as an in-place upgrade.

Each environment can also be managed directly:

```sh
terraform -chdir=terraform/envs/local-dev init
terraform -chdir=terraform/envs/local-dev plan
terraform -chdir=terraform/envs/local-dev apply
```

Review the plan before applying. `make up` is a convenience wrapper that applies all three environments.
