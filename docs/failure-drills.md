# Failure drill evidence

Use the postmortem template after completing each drill. Capture the linked alert, diagnostic commands, fix, and timestamps from the actual run. This file tracks planned drills and does not claim they have been executed.

| Drill | Injection | Detection | Runbook | Evidence |
|---|---|---|---|---|
| Bad image tag | Set a non-existent sample-app tag in a dev PR | Argo CD Degraded, `DeploymentReplicasMismatch` | [rollback-bad-deploy.md](runbooks/rollback-bad-deploy.md) | Pending live run |
| Node failure | Stop one dev worker container | `NodeNotReady`, `fleetctl audit` | [node-not-ready.md](runbooks/node-not-ready.md) | Pending live run |
| Out of memory | Lower the dev app memory limit to 10Mi in Git | `PodCrashLooping`, OOMKilled events | [rollback-bad-deploy.md](runbooks/rollback-bad-deploy.md) | Pending live run |
| DNS outage | Scale dev CoreDNS to zero | app errors, `SampleAppHighErrorRate` | [rollback-bad-deploy.md](runbooks/rollback-bad-deploy.md) | Pending live run |
