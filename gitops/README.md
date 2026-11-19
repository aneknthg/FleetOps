# GitOps structure

Argo CD runs on the `mgmt` cluster. The root Application under `gitops/bootstrap/` manages the ApplicationSet and the dev monitoring Helm Application. The ApplicationSet creates one app per cluster/app overlay; dev auto-syncs with self-heal while prod is manual.

The workload app at `gitops/apps/sample-app/` deploys the sample service. Monitoring is split into a chart Application in `gitops/bootstrap/apps/monitoring-stack-dev.yaml` and Kubernetes resources in `gitops/apps/monitoring/overlays/dev/`. This keeps Argo CD Application objects on the management cluster and the service monitor, alert rules, and dashboard on dev.

The sample service exposes `/healthz`, `/version`, and `/metrics`. Prometheus scrapes it through a ServiceMonitor; four alert rules link to runbooks. Grafana discovers the labeled dashboard ConfigMap through its sidecar.

## Build and publish the app image

```sh
docker build -t ghcr.io/aneknthg/fleetops-sample-app:0.1.0 ./sample-app
docker push ghcr.io/aneknthg/fleetops-sample-app:0.1.0
```

Authenticate to GHCR before pushing. Never put a registry token or kubeconfig in Git. The app deployment uses the GHCR image; Argo CD can pull it once the package is public or image pull credentials are configured.

## Bootstrap

After `make up`, run `scripts/bootstrap.sh`. The script installs the pinned Argo CD Helm chart on `mgmt`, registers the workload clusters using kind's internal endpoints, and applies the root Application. It requires Docker, kind, Helm, kubectl, and argocd CLIs.
