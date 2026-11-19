# fleetctl

`fleetctl` operates through kubeconfig contexts using `client-go`. Every Kubernetes request has a bounded timeout. `audit --all` continues when a context is unreachable and reports it as a high-severity finding.

## Build and test

```sh
go test ./...
go build -o fleetctl .
```

## Commands

### `fleetctl clusters`

Lists each kubeconfig context, reachability, API server version, and node count.

```text
CONTEXT                         REACHABLE  API VERSION  NODES  ERROR
kind-dev                        true       v1.36.6      3
kind-prod                       true       v1.36.6      3
```

### `fleetctl nodes --context kind-dev [-o json]`

Lists node readiness, kubelet version, OS image, kernel, age, and minor-version skew from the control plane. Use `-o json` for machine-readable output.

### `fleetctl images --context kind-dev [-o json]`

Lists unique container images across regular, init, and ephemeral containers, with the number of pods using each image.

### `fleetctl audit [--all] [--context kind-dev]`

Runs seven rules covering version skew, kernel age, node and pod health, mutable image tags, PDB coverage, and resource limits. `--all` evaluates contexts concurrently. The command exits with status 1 if it finds any high-severity issue.

Set the `minimumKernelVersion` in `fleetctl.yaml` to change the medium-severity kernel threshold.

### `fleetctl drain NODE --context kind-dev [--yes] [--timeout 5m]`

Dry-run is the default. The command reports pods that would be evicted. With `--yes`, it cordons the node and evicts pods using the Kubernetes Eviction API, which honors PodDisruptionBudgets. DaemonSet and static mirror pods are skipped. It waits for workload pods to leave and prints the exact `kubectl uncordon` command. If an eviction is rejected, the node remains cordoned and the error includes the uncordon command.

```sh
fleetctl drain dev-worker --context kind-dev
fleetctl drain dev-worker --context kind-dev --yes --timeout 10m
```

The drain path is covered with client-go fake-client tests, including a rejected eviction.
