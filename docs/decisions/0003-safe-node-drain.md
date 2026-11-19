# ADR 0003: Drain nodes through the Eviction API

- Status: Accepted
- Date: 2026-11-18

## Context

Node maintenance must respect workload disruption budgets and provide an operator a safe preview.

## Decision

`fleetctl drain` is dry-run by default. When confirmed, it cordons the node and requests pod evictions through the Kubernetes Eviction API, skipping DaemonSet and static mirror pods.

## Consequences

PDBs can reject eviction and leave the node cordoned for operator review. The CLI reports the explicit uncordon command and never silently forces deletion.
