# ADR 0001: Use kind for local multi-cluster practice

- Status: Accepted
- Date: 2026-11-18

## Context

The lab needs repeatable management, dev, and prod clusters without ongoing cloud charges. The local environment has no AWS credentials or approval to create billable resources.

## Decision

Use Terraform with the `tehcyx/kind` provider for three local clusters. Keep AWS EKS as a separate, explicitly applied and destroyed exercise.

## Consequences

Local clusters are quick to recreate and support GitOps experiments. They share the host kernel and do not prove cloud networking, IAM, autoscaling, or large fleet operations.
