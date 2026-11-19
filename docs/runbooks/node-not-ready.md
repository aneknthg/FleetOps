# Node Not Ready runbook

## Symptoms

Prometheus fires `NodeNotReady`, `fleetctl audit` reports a high-severity node finding, or a node disappears from `kubectl get nodes`.

## Impact

Pods on the node may stop serving traffic or reschedule elsewhere. A PodDisruptionBudget can limit voluntary movement.

## Diagnosis

```sh
kubectl --context kind-dev get nodes -o wide
kubectl --context kind-dev describe node <node>
kubectl --context kind-dev get pods -A -o wide --field-selector spec.nodeName=<node>
fleetctl nodes --context kind-dev
fleetctl audit --context kind-dev
```

Check node conditions, kubelet events, Docker container state, and whether the remaining workers have capacity.

## Fix

For the kind failure drill, restart the stopped node container with `docker start <container>` and wait for Ready. For a persistent failure, drain one node at a time through `fleetctl drain <node> --context kind-dev --yes`, repair or replace it, then verify node and workload health.

## Rollback

If a repair makes health worse, stop changing other nodes, keep the affected node cordoned, restore the previous node image or container state, and verify PDBs before retrying.

## Escalation

Escalate if the control plane is unreachable, multiple nodes are NotReady, PDBs block recovery, or workloads cannot fit on remaining capacity.
