package audit

import (
	"testing"
	"time"
)

func TestAuditRules(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		run  func() []Finding
		want int
	}{
		{
			name: "kubelet minor skew",
			run: func() []Finding { return RuleKubeletSkew(Snapshot{ControlPlaneVersion:"v1.36.6",Nodes:[]Node{{Name:"same",KubeletVersion:"v1.36.5"},{Name:"old",KubeletVersion:"v1.35.9"}}}) },
			want: 1,
		},
		{
			name: "kernel below minimum",
			run: func() []Finding { return RuleMinimumKernel(Snapshot{Nodes:[]Node{{Name:"old",KernelVersion:"5.10.0-20"},{Name:"new",KernelVersion:"5.15.0-1"}}},"5.15.0") },
			want: 1,
		},
		{
			name: "node readiness and pressure",
			run: func() []Finding { return RuleNodeHealth(Snapshot{Nodes:[]Node{{Name:"bad",Ready:false,Pressure:true}}}) },
			want: 2,
		},
		{
			name: "old non-ready and crash loop pods",
			run: func() []Finding { return RulePodHealth(Snapshot{Pods:[]Pod{{Namespace:"default",Name:"crash",Waiting:"CrashLoopBackOff"},{Namespace:"default",Name:"pending",CreatedAt:now.Add(-6*time.Minute),Ready:false}}},now) },
			want: 2,
		},
		{
			name: "latest and untagged images",
			run: func() []Finding { return RuleImageTags(Snapshot{Pods:[]Pod{{Namespace:"default",Name:"p",Containers:[]Container{{Name:"latest",Image:"nginx:latest"},{Name:"untagged",Image:"nginx"},{Name:"pinned",Image:"nginx@sha256:abcd"}}}}}) },
			want: 2,
		},
		{
			name: "replicated deployment without matching PDB",
			run: func() []Finding {
				s := Snapshot{Deployments:[]Deployment{{Namespace:"default",Name:"app",Replicas:2,Selector:map[string]string{"app":"a"}}},PDBs:[]PDB{{Namespace:"default",Name:"other",Selector:map[string]string{"app":"b"}}}}
				return RuleDeploymentPDB(s)
			},
			want: 1,
		},
		{
			name: "container without limits",
			run: func() []Finding { return RuleResourceLimits(Snapshot{Pods:[]Pod{{Namespace:"default",Name:"p",Containers:[]Container{{Name:"limited",HasLimits:true},{Name:"unlimited"}}}}}) },
			want: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := len(test.run()); got != test.want {
				t.Fatalf("got %d finding(s), want %d", got, test.want)
			}
		})
	}
}
