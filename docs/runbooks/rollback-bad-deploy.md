# Roll back a bad deployment

## Symptoms

Argo CD reports an application as Degraded or Progressing, `DeploymentReplicasMismatch` fires, pods show ImagePullBackOff or CrashLoopBackOff, or `SampleAppHighErrorRate` exceeds 5 percent.

## Impact

The affected service may have reduced availability or return 5xx responses. Check the replica count and error rate before making further changes.

## Diagnosis

```sh
kubectl --context kind-dev -n default get deploy,pods,events
kubectl --context kind-dev -n default describe deployment sample-app
kubectl --context kind-dev -n default logs deploy/sample-app --all-containers
argocd app get sample-app-dev
```

Identify the Git commit and image tag that introduced the failure. Check Argo CD sync status and the alert's linked runbook.

## Fix

Revert the bad Git change in a pull request. Argo CD will reconcile the previous desired image and settings. Confirm pods become Ready and the error rate returns below threshold.

## Rollback

If the revert itself is incorrect, revert that revert in a new pull request and pause further rollout until the image or manifest is understood.

## Escalation

Escalate if rollback does not restore service, the previous image is unavailable, data integrity may be affected, or several clusters are impacted.
