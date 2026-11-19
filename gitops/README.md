# Sample app and GitOps

The sample service exposes `/healthz` and `/version`. Build it with:

```sh
docker build -t FleetOps-sample-app:0.1.0 ./sample-app
```

Before bootstrapping, replace `aneknthg` in the GitOps manifests with the GitHub owner and push the image to GHCR. Do not commit a registry token.

`gitops/apps/sample-app/base` defines two replicas, resource requests and limits, health probes, a PDB, and topology spreading. The `dev` overlay is intended to auto-sync; `prod` is intended to sync manually. `gitops/bootstrap/applicationset.yaml` generates an Application for each app overlay and registered cluster.

The bootstrap script installs the Argo CD Helm chart version 10.18.0 on `mgmt`, registers `dev` and `prod` using internal kind endpoints, then applies the root Application. It requires local clusters named `mgmt`, `dev`, and `prod`, and the `argocd`, `helm`, and `kind` CLIs.
