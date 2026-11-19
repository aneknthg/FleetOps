package k8s

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestDrainDryRunDoesNotChangeCluster(t *testing.T) {
	node := &corev1.Node{ObjectMeta:metav1.ObjectMeta{Name:"worker-1"}}
	pod := &corev1.Pod{ObjectMeta:metav1.ObjectMeta{Name:"app-1",Namespace:"default"},Spec:corev1.PodSpec{NodeName:"worker-1"}}
	client:=fake.NewSimpleClientset(node,pod)
	pods,err:=DrainNode(context.Background(),client,"worker-1",true,time.Minute)
	if err!=nil{t.Fatal(err)}
	if len(pods)!=1 || pods[0]!="default/app-1"{t.Fatalf("unexpected dry-run pods: %#v",pods)}
	updated,err:=client.CoreV1().Nodes().Get(context.Background(),"worker-1",metav1.GetOptions{})
	if err!=nil{t.Fatal(err)}
	if updated.Spec.Unschedulable{t.Fatal("dry-run unexpectedly cordoned node")}
}

func TestDrainDoesNotBypassDisruptionBudget(t *testing.T) {
	node:=&corev1.Node{ObjectMeta:metav1.ObjectMeta{Name:"worker-1"}}
	pod:=&corev1.Pod{ObjectMeta:metav1.ObjectMeta{Name:"app-1",Namespace:"default"},Spec:corev1.PodSpec{NodeName:"worker-1"}}
	client:=fake.NewSimpleClientset(node,pod)
	client.PrependReactor("create","pods",func(action ktesting.Action)(bool,runtime.Object,error){
		if action.GetSubresource()=="eviction"{return true,nil,apierrors.NewTooManyRequests("blocked by PodDisruptionBudget",1)}
		return false,nil,nil
	})
	_,err:=DrainNode(context.Background(),client,"worker-1",false,time.Second)
	if err==nil{t.Fatal("expected eviction rejection from PodDisruptionBudget")}
	updated,getErr:=client.CoreV1().Nodes().Get(context.Background(),"worker-1",metav1.GetOptions{})
	if getErr!=nil{t.Fatal(getErr)}
	if !updated.Spec.Unschedulable{t.Fatal("node should remain cordoned after rejected eviction")}
}
