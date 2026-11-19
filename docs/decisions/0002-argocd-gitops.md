# ADR 0002: Use Argo CD for GitOps

- Status: Accepted
- Date: 2026-11-18

## Context

The lab needs a reviewable, repeatable delivery flow across dev and prod with drift correction and rollback history.

## Decision

Use Argo CD with a root Application and ApplicationSet to discover per-app cluster overlays. Dev auto-syncs with self-heal; prod receives a generated Application without automated sync.

## Consequences

Desired state is reviewed in Git and reconciled continuously. Cluster registration and the management cluster remain bootstrap dependencies; this lab does not claim production-grade multi-tenant policy or fleet scale.
