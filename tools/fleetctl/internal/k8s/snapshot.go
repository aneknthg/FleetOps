package k8s

import (
	"context"
	"time"

	"github.com/aneknthg/fleetops/tools/fleetctl/internal/audit"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func Snapshot(ctx context.Context, client kubernetes.Interface, cluster string) (audit.Snapshot, error) {
	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		return audit.Snapshot{}, err
	}
	snapshot := audit.Snapshot{Cluster:cluster,ControlPlaneVersion:serverVersion.GitVersion}
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1ListOptions())
	if err != nil { return audit.Snapshot{}, err }
	for _, node := range nodes.Items {
		entry:=audit.Node{Name:node.Name,KubeletVersion:node.Status.NodeInfo.KubeletVersion,KernelVersion:node.Status.NodeInfo.KernelVersion}
		for _, condition:=range node.Status.Conditions {
			if condition.Type==corev1.NodeReady { entry.Ready=condition.Status==corev1.ConditionTrue }
			if (condition.Type==corev1.NodeMemoryPressure || condition.Type==corev1.NodeDiskPressure || condition.Type==corev1.NodePIDPressure) && condition.Status==corev1.ConditionTrue { entry.Pressure=true }
		}
		snapshot.Nodes=append(snapshot.Nodes,entry)
	}
	pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1ListOptions())
	if err != nil { return audit.Snapshot{}, err }
	for _, pod := range pods.Items {
		entry:=audit.Pod{Namespace:pod.Namespace,Name:pod.Name,CreatedAt:pod.CreationTimestamp.Time}
		for _, condition:=range pod.Status.Conditions { if condition.Type==corev1.PodReady { entry.Ready=condition.Status==corev1.ConditionTrue } }
		for _, status:=range append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...) { if status.State.Waiting!=nil && entry.Waiting=="" { entry.Waiting=status.State.Waiting.Reason } }
		for _, container:=range pod.Spec.Containers { entry.Containers=append(entry.Containers,audit.Container{Name:container.Name,Image:container.Image,HasResources:len(container.Resources.Requests)>0 || len(container.Resources.Limits)>0,HasLimits:len(container.Resources.Limits)>0}) }
		for _, container:=range pod.Spec.InitContainers { entry.Containers=append(entry.Containers,audit.Container{Name:container.Name,Image:container.Image,HasResources:len(container.Resources.Requests)>0 || len(container.Resources.Limits)>0,HasLimits:len(container.Resources.Limits)>0}) }
		for _, container:=range pod.Spec.EphemeralContainers { entry.Containers=append(entry.Containers,audit.Container{Name:container.Name,Image:container.Image}) }
		snapshot.Pods=append(snapshot.Pods,entry)
	}
	deployments, err := client.AppsV1().Deployments(metav1.NamespaceAll).List(ctx, metav1ListOptions())
	if err != nil { return audit.Snapshot{}, err }
	for _, deployment:=range deployments.Items { snapshot.Deployments=append(snapshot.Deployments,audit.Deployment{Namespace:deployment.Namespace,Name:deployment.Name,Replicas:replicaCount(deployment.Spec.Replicas),Selector:deployment.Spec.Selector.MatchLabels}) }
	budgets, err := client.PolicyV1().PodDisruptionBudgets(metav1.NamespaceAll).List(ctx, metav1ListOptions())
	if err != nil { return audit.Snapshot{}, err }
	for _, budget:=range budgets.Items { selector:=map[string]string{}; if budget.Spec.Selector!=nil { selector=budget.Spec.Selector.MatchLabels }; snapshot.PDBs=append(snapshot.PDBs,audit.PDB{Namespace:budget.Namespace,Name:budget.Name,Selector:selector}) }
	return snapshot,nil
}

func EvictPod(ctx context.Context, client kubernetes.Interface, pod *corev1.Pod) error {
	grace := int64(30)
	return client.CoreV1().Pods(pod.Namespace).EvictV1(ctx,&policyv1.Eviction{ObjectMeta:metav1.ObjectMeta{Name:pod.Name,Namespace:pod.Namespace},DeleteOptions:&metav1.DeleteOptions{GracePeriodSeconds:&grace}})
}

func IsDaemonSetPod(pod corev1.Pod) bool { for _, owner:=range pod.OwnerReferences { if owner.Kind=="DaemonSet" { return true } }; return false }
func PodSelector(pod corev1.Pod) labels.Set { return labels.Set(pod.Labels) }
func Now() time.Time { return time.Now() }

func replicaCount(replicas *int32) int32 {
	if replicas == nil {
		return 1
	}
	return *replicas
}



