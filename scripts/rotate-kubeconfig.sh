#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "Usage: $0 <mgmt|dev|prod> [--internal]" >&2
  exit 2
fi

readonly cluster="$1"
internal_flag="${2:-}"
case "$cluster" in
  mgmt|dev|prod) ;;
  *) echo "Cluster must be mgmt, dev, or prod." >&2; exit 2 ;;
esac
if [[ -n "$internal_flag" && "$internal_flag" != "--internal" ]]; then
  echo "The only supported option is --internal." >&2
  exit 2
fi

for command in kind chmod; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "Required command not found: $command" >&2
    exit 127
  fi
done

kubeconfig_path="${KUBECONFIG:-${HOME}/.kube/config}"
mkdir -p "$(dirname "$kubeconfig_path")"
kind_args=(export kubeconfig --name "$cluster" --kubeconfig "$kubeconfig_path")
if [[ "$internal_flag" == "--internal" ]]; then
  kind_args+=(--internal)
fi
kind "${kind_args[@]}"
chmod 600 "$kubeconfig_path"
printf 'Updated kubeconfig: %s\n' "$kubeconfig_path"
printf 'Check access with: kubectl --context kind-%s get nodes\n' "$cluster"
