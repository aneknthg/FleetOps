package audit

import "time"

type Severity string

const (
	SeverityHigh   Severity = "HIGH"
	SeverityMedium Severity = "MEDIUM"
	SeverityLow    Severity = "LOW"
)

type Finding struct {
	Cluster  string   `json:"cluster"`
	Severity Severity `json:"severity"`
	Rule     string   `json:"rule"`
	Resource string   `json:"resource"`
	Message  string   `json:"message"`
}

type Snapshot struct {
	Cluster             string
	ControlPlaneVersion string
	Nodes               []Node
	Pods                []Pod
	Deployments         []Deployment
	PDBs                []PDB
}

type Node struct {
	Name          string
	Ready         bool
	Pressure      bool
	KubeletVersion string
	KernelVersion string
}

type Container struct {
	Name         string
	Image        string
	HasResources bool
	HasLimits    bool
}

type Pod struct {
	Namespace string
	Name      string
	Ready     bool
	CreatedAt time.Time
	Waiting   string
	Containers []Container
}

type Deployment struct {
	Namespace string
	Name      string
	Replicas  int32
	Selector  map[string]string
}

type PDB struct {
	Namespace string
	Name      string
	Selector  map[string]string
}
