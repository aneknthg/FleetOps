package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

func DrainNode(ctx context.Context, client kubernetes.Interface, nodeName string, dryRun bool, timeout time.Duration) ([]string, error) {
	_, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get node %s: %w", nodeName, err)
	}
	pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{FieldSelector:"spec.nodeName="+nodeName})
	if err != nil {
		return nil, fmt.Errorf("list pods on node %s: %w", nodeName, err)
	}
	targets := make([]corev1.Pod, 0, len(pods.Items))
	for _, pod := range pods.Items {
		if IsDaemonSetPod(pod) || pod.Annotations[corev1.MirrorPodAnnotationKey] != "" {
			continue
		}
		targets = append(targets, pod)
	}
	podNames := make([]string, 0, len(targets))
	for _, pod := range targets {
		podNames = append(podNames, pod.Namespace+"/"+pod.Name)
	}
	if dryRun {
		return podNames, nil
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current, getErr := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		if getErr != nil { return getErr }
		current.Spec.Unschedulable = true
		_, updateErr := client.CoreV1().Nodes().Update(ctx, current, metav1.UpdateOptions{})
		return updateErr
	}); err != nil {
		return podNames, fmt.Errorf("cordon node %s: %w", nodeName, err)
	}
	for i := range targets {
		if err := EvictPod(ctx, client, &targets[i]); err != nil {
			return podNames, fmt.Errorf("evict %s/%s (node remains cordoned; run kubectl uncordon %s): %w", targets[i].Namespace, targets[i].Name, nodeName, err)
		}
	}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		remaining, listErr := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{FieldSelector:"spec.nodeName="+nodeName})
		if listErr != nil { return podNames, fmt.Errorf("wait for pods to leave node: %w", listErr) }
		active := 0
		for _, pod := range remaining.Items { if !IsDaemonSetPod(pod) && pod.Annotations[corev1.MirrorPodAnnotationKey] == "" { active++ } }
		if active == 0 { return podNames, nil }
		select {
		case <-ctx.Done(): return podNames, ctx.Err()
		case <-deadline.C: return podNames, fmt.Errorf("timed out waiting for %d pod(s); node remains cordoned; run kubectl uncordon %s", active, nodeName)
		case <-ticker.C:
		}
	}
}

