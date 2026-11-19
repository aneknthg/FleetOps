#!/usr/bin/env bash
set -euo pipefail

readonly argocd_namespace="argocd"
readonly argocd_chart_version="10.18.0"

for command in kubectl helm kind argocd; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "Required command not found: $command" >&2
    exit 127
  fi
done

if grep -R -n "OWNER" gitops/bootstrap gitops/apps; then
  echo "Replace OWNER in GitOps manifests with the actual GitHub owner before bootstrapping." >&2
  exit 2
fi

kubectl --context kind-mgmt create namespace "$argocd_namespace" \
  --dry-run=client -o yaml | kubectl --context kind-mgmt apply -f -
helm repo add argo https://argoproj.github.io/argo-helm --force-update
helm upgrade --install argocd argo/argo-cd \
  --version "$argocd_chart_version" \
  --kube-context kind-mgmt \
  --namespace "$argocd_namespace" \
  --create-namespace \
  --set configs.params."server\.insecure"=true

for cluster in dev prod; do
  kind export kubeconfig --name "$cluster" --internal
  argocd --core --kube-context kind-mgmt cluster add "kind-$cluster" \
    --cluster-endpoint internal \
    --name "$cluster" \
    --upsert \
    --yes
done

kubectl --context kind-mgmt apply -f gitops/bootstrap/root-app.yaml
