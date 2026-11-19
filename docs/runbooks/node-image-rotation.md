# Node image rotation runbook

## Symptoms

`fleetctl audit` reports a node kernel below `minimumKernelVersion`, a node OS image is out of policy, or a planned Kubernetes node image upgrade is required.

## Impact

A node replacement temporarily reduces schedulable capacity. PDBs may prevent eviction, so rotate one worker at a time and confirm spare capacity before proceeding.

## Diagnosis

```sh
fleetctl nodes --context kind-dev
fleetctl audit --context kind-dev
kubectl --context kind-dev get nodes -o wide
kubectl --context kind-dev get pdb -A
```

Record the current Terraform node image, Kubernetes version, kernel, node count, and workload health. kind clusters share the host kernel and cannot model an independent kernel upgrade.

## Fix

1. Update the pinned `node_image` in `terraform/modules/kind-cluster/variables.tf` through a pull request.
2. Review `terraform plan`; changing kind node images replaces the cluster, so use a non-production local environment for this exercise.
3. For EKS, update the Kubernetes version in the EKS environment/module and review the plan. Apply the control-plane update first, then rotate the managed node group one node at a time.
4. Drain with `fleetctl drain <node> --context <context> --yes --timeout 10m`, replace the node, and wait for Ready and workload health before proceeding to the next node.
5. Update `minimumKernelVersion` only after confirming the intended fleet baseline.

## Rollback

For a failed kind replacement, restore the prior node image and recreate from the known-good Terraform configuration. For EKS, follow the AWS-supported version-skew policy; do not downgrade the control plane. Restore a previous node AMI or node group configuration only when AWS supports it.

## Escalation

Escalate if PDBs prevent safe evacuation, the control plane becomes unhealthy, nodes fail to join, or the service has insufficient capacity to tolerate one node at a time.
