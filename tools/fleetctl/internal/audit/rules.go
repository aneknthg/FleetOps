package audit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	versionPattern = regexp.MustCompile(`(^|[^0-9])v?1\.([0-9]+)(\.[0-9]+)?`)
	kernelPattern  = regexp.MustCompile(`([0-9]+\.[0-9]+\.[0-9]+)`)
)

func Analyze(s Snapshot, minimumKernel string, now time.Time) []Finding {
	findings := make([]Finding, 0)
	findings = append(findings, RuleKubeletSkew(s)...)
	findings = append(findings, RuleMinimumKernel(s, minimumKernel)...)
	findings = append(findings, RuleNodeHealth(s)...)
	findings = append(findings, RulePodHealth(s, now)...)
	findings = append(findings, RuleImageTags(s)...)
	findings = append(findings, RuleDeploymentPDB(s)...)
	findings = append(findings, RuleResourceLimits(s)...)
	return findings
}

func RuleKubeletSkew(s Snapshot) []Finding {
	cpMinor, ok := minorVersion(s.ControlPlaneVersion)
	if !ok {
		return nil
	}
	var findings []Finding
	for _, node := range s.Nodes {
		minor, ok := minorVersion(node.KubeletVersion)
		if ok && minor != cpMinor {
			findings = append(findings, finding(s, SeverityHigh, "node-version-skew", node.Name,
				fmt.Sprintf("kubelet %s differs in minor version from control plane %s", node.KubeletVersion, s.ControlPlaneVersion)))
		}
	}
	return findings
}

func RuleMinimumKernel(s Snapshot, minimum string) []Finding {
	if minimum == "" {
		return nil
	}
	minParts, ok := versionParts(minimum)
	if !ok {
		return []Finding{finding(s, SeverityLow, "invalid-minimum-kernel", "fleetctl.yaml", "minimumKernelVersion is not a numeric dotted version")}
	}
	var findings []Finding
	for _, node := range s.Nodes {
		current, ok := versionParts(node.KernelVersion)
		if ok && compareParts(current, minParts) < 0 {
			findings = append(findings, finding(s, SeverityMedium, "old-kernel", node.Name,
				fmt.Sprintf("kernel %s is older than configured minimum %s", node.KernelVersion, minimum)))
		}
	}
	return findings
}

func RuleNodeHealth(s Snapshot) []Finding {
	var findings []Finding
	for _, node := range s.Nodes {
		if !node.Ready {
			findings = append(findings, finding(s, SeverityHigh, "node-not-ready", node.Name, "node is not Ready"))
		}
		if node.Pressure {
			findings = append(findings, finding(s, SeverityHigh, "node-pressure", node.Name, "node reports a pressure condition"))
		}
	}
	return findings
}

func RulePodHealth(s Snapshot, now time.Time) []Finding {
	var findings []Finding
	for _, pod := range s.Pods {
		resource := pod.Namespace + "/" + pod.Name
		if pod.Waiting == "CrashLoopBackOff" {
			findings = append(findings, finding(s, SeverityMedium, "pod-crash-loop", resource, "container is in CrashLoopBackOff"))
			continue
		}
		if !pod.Ready && now.Sub(pod.CreatedAt) > 5*time.Minute {
			findings = append(findings, finding(s, SeverityMedium, "pod-not-ready", resource, "pod has not been Ready for more than 5 minutes"))
		}
	}
	return findings
}

func RuleImageTags(s Snapshot) []Finding {
	var findings []Finding
	for _, pod := range s.Pods {
		for _, container := range pod.Containers {
			image := container.Image
			if imageNeedsTag(image) {
				findings = append(findings, finding(s, SeverityMedium, "mutable-image-tag", pod.Namespace+"/"+pod.Name,
					fmt.Sprintf("container %s uses image %s without an explicit non-latest tag", container.Name, image)))
			}
		}
	}
	return findings
}

func RuleDeploymentPDB(s Snapshot) []Finding {
	var findings []Finding
	for _, deployment := range s.Deployments {
		if deployment.Replicas <= 1 {
			continue
		}
		matched := false
		for _, pdb := range s.PDBs {
			if pdb.Namespace == deployment.Namespace && selectorMatches(pdb.Selector, deployment.Selector) {
				matched = true
				break
			}
		}
		if !matched {
			resource := deployment.Namespace + "/" + deployment.Name
			findings = append(findings, finding(s, SeverityLow, "missing-pdb", resource,
				"deployment has more than one replica and no matching PodDisruptionBudget"))
		}
	}
	return findings
}

func RuleResourceLimits(s Snapshot) []Finding {
	var findings []Finding
	for _, pod := range s.Pods {
		for _, container := range pod.Containers {
			if !container.HasLimits {
				resource := pod.Namespace + "/" + pod.Name
				findings = append(findings, finding(s, SeverityLow, "missing-resource-limits", resource,
					"container "+container.Name+" has no resource limits"))
			}
		}
	}
	return findings
}

func imageNeedsTag(image string) bool {
	if image == "" {
		return true
	}
	if strings.Contains(image, "@sha256:") {
		return false
	}
	last := image[strings.LastIndex(image, "/")+1:]
	colon := strings.LastIndex(last, ":")
	if colon < 0 {
		return true
	}
	return last[colon+1:] == "latest" || last[colon+1:] == ""
}
func selectorMatches(selector, labels map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for key, value := range selector {
		if labels[key] != value {
			return false
		}
	}
	return true
}

func minorVersion(version string) (int, bool) {
	match := versionPattern.FindStringSubmatch(version)
	if len(match) != 3 {
		return 0, false
	}
	minor, err := strconv.Atoi(match[2])
	return minor, err == nil
}

func versionParts(version string) ([]int, bool) {
	match := kernelPattern.FindStringSubmatch(version)
	if len(match) != 2 {
		return nil, false
	}
	parts := strings.Split(match[1], ".")
	result := make([]int, len(parts))
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, false
		}
		result[i] = value
	}
	return result, true
}
func compareParts(a, b []int) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] { return -1 }
		if a[i] > b[i] { return 1 }
	}
	return 0
}

func finding(s Snapshot, severity Severity, rule, resource, message string) Finding {
	return Finding{Cluster: s.Cluster, Severity: severity, Rule: rule, Resource: resource, Message: message}
}









