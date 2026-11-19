package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/aneknthg/fleetops/tools/fleetctl/internal/audit"
	fleetk8s "github.com/aneknthg/fleetops/tools/fleetctl/internal/k8s"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
)

var ErrHighFindings = errors.New("high-severity audit findings found")

var kubeconfigPath string

func Execute() error { return NewRootCmd().Execute() }

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "fleetctl",
		Short:         "Inspect and operate a Kubernetes cluster fleet",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&kubeconfigPath, "kubeconfig", fleetk8s.DefaultKubeconfigPath(), "Path to kubeconfig")
	root.AddCommand(newClustersCommand(), newNodesCommand(), newImagesCommand(), newAuditCommand(), newDrainCommand())
	return root
}

type clusterRow struct {
	Context string `json:"context"`
	Reachable bool `json:"reachable"`
	Version string `json:"version,omitempty"`
	NodeCount int `json:"nodeCount"`
	Error string `json:"error,omitempty"`
}

func newClustersCommand() *cobra.Command {
	return &cobra.Command{
		Use: "clusters",
		Short: "List kubeconfig contexts and their cluster health",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := fleetk8s.LoadKubeconfig(kubeconfigPath)
			if err != nil { return err }
			contexts := sortedContextNames(config.Contexts)
			rows := make([]clusterRow, len(contexts))
			g, ctx := errgroup.WithContext(cmd.Context())
			g.SetLimit(8)
			for i, name := range contexts {
				i, name := i, name
				g.Go(func() error {
					row := clusterRow{Context:name}
					client, clientErr := fleetk8s.NewClient(kubeconfigPath,name)
					if clientErr != nil { row.Error=clientErr.Error(); rows[i]=row; return nil }
					requestCtx, cancel := context.WithTimeout(ctx,10*time.Second); defer cancel()
					version, versionErr := client.Discovery().ServerVersion()
					if versionErr != nil { row.Error=versionErr.Error(); rows[i]=row; return nil }
					nodes, listErr := client.CoreV1().Nodes().List(requestCtx,metav1.ListOptions{})
					if listErr != nil { row.Error=listErr.Error(); rows[i]=row; return nil }
					row.Reachable=true; row.Version=version.GitVersion; row.NodeCount=len(nodes.Items); rows[i]=row
					return nil
				})
			}
			if err:=g.Wait();err!=nil{return err}
			return writeClusterRows(cmd.OutOrStdout(),rows)
		},
	}
}

func writeClusterRows(w io.Writer, rows []clusterRow) error {
	tw:=tabwriter.NewWriter(w,0,4,2,' ',0)
	fmt.Fprintln(tw,"CONTEXT\tREACHABLE\tAPI VERSION\tNODES\tERROR")
	for _,row:=range rows { fmt.Fprintf(tw,"%s\t%t\t%s\t%d\t%s\n",row.Context,row.Reachable,row.Version,row.NodeCount,row.Error) }
	return tw.Flush()
}

type nodeRow struct {
	Name string `json:"name"`
	Ready bool `json:"ready"`
	KubeletVersion string `json:"kubeletVersion"`
	OSImage string `json:"osImage"`
	KernelVersion string `json:"kernelVersion"`
	Age string `json:"age"`
	VersionSkew bool `json:"versionSkew"`
}

func newNodesCommand() *cobra.Command {
	var contextName, output string
	command:=&cobra.Command{
		Use:"nodes",
		Short:"List node health and version details",
		Args:cobra.NoArgs,
		RunE:func(cmd *cobra.Command,_ []string)error{
			if contextName=="" { return errors.New("--context is required") }
			client,err:=fleetk8s.NewClient(kubeconfigPath,contextName);if err!=nil{return err}
			ctx,cancel:=context.WithTimeout(cmd.Context(),20*time.Second);defer cancel()
			server,err:=client.Discovery().ServerVersion();if err!=nil{return err}
			nodes,err:=client.CoreV1().Nodes().List(ctx,metav1.ListOptions{});if err!=nil{return err}
			rows:=make([]nodeRow,0,len(nodes.Items))
			for _,node:=range nodes.Items{
				ready:=false
				for _,condition:=range node.Status.Conditions{if condition.Type==corev1.NodeReady{ready=condition.Status==corev1.ConditionTrue}}
				rows=append(rows,nodeRow{Name:node.Name,Ready:ready,KubeletVersion:node.Status.NodeInfo.KubeletVersion,OSImage:node.Status.NodeInfo.OSImage,KernelVersion:node.Status.NodeInfo.KernelVersion,Age:time.Since(node.CreationTimestamp.Time).Round(time.Second).String(),VersionSkew:minor(node.Status.NodeInfo.KubeletVersion)!=minor(server.GitVersion)})
			}
			return writeOutput(cmd.OutOrStdout(),output,rows,func(w io.Writer)error{
				tw:=tabwriter.NewWriter(w,0,4,2,' ',0);fmt.Fprintln(tw,"NAME\tREADY\tKUBELET\tOS IMAGE\tKERNEL\tAGE\tVERSION SKEW")
				for _,row:=range rows{fmt.Fprintf(tw,"%s\t%t\t%s\t%s\t%s\t%s\t%t\n",row.Name,row.Ready,row.KubeletVersion,row.OSImage,row.KernelVersion,row.Age,row.VersionSkew)};return tw.Flush()
			})
		},
	}
	command.Flags().StringVarP(&contextName,"context","c","","Kubeconfig context")
	command.Flags().StringVarP(&output,"output","o","table","Output format: table or json")
	return command
}

type imageRow struct { Image string `json:"image"`; PodCount int `json:"podCount"` }

func newImagesCommand() *cobra.Command {
	var contextName, output string
	command:=&cobra.Command{Use:"images",Short:"List unique pod images and counts",Args:cobra.NoArgs,RunE:func(cmd *cobra.Command,_ []string)error{
		if contextName==""{return errors.New("--context is required")}
		client,err:=fleetk8s.NewClient(kubeconfigPath,contextName);if err!=nil{return err}
		ctx,cancel:=context.WithTimeout(cmd.Context(),20*time.Second);defer cancel()
		pods,err:=client.CoreV1().Pods(metav1.NamespaceAll).List(ctx,metav1.ListOptions{});if err!=nil{return err}
		counts:=map[string]int{}
		for _,pod:=range pods.Items{
			seen:=map[string]struct{}{}
			for _,container:=range pod.Spec.Containers{seen[container.Image]=struct{}{}}
			for _,container:=range pod.Spec.InitContainers{seen[container.Image]=struct{}{}}
			for _,container:=range pod.Spec.EphemeralContainers{seen[container.Image]=struct{}{}}
			for image:=range seen{counts[image]++}
		}
		images:=make([]string,0,len(counts));for image:=range counts{images=append(images,image)};sort.Strings(images)
		rows:=make([]imageRow,0,len(images));for _,image:=range images{rows=append(rows,imageRow{Image:image,PodCount:counts[image]})}
		return writeOutput(cmd.OutOrStdout(),output,rows,func(w io.Writer)error{tw:=tabwriter.NewWriter(w,0,4,2,' ',0);fmt.Fprintln(tw,"IMAGE\tPODS");for _,row:=range rows{fmt.Fprintf(tw,"%s\t%d\n",row.Image,row.PodCount)};return tw.Flush()})
	}}
	command.Flags().StringVarP(&contextName,"context","c","","Kubeconfig context")
	command.Flags().StringVarP(&output,"output","o","table","Output format: table or json")
	return command
}

func newAuditCommand() *cobra.Command {
	var contextName string
	var all bool
	command:=&cobra.Command{Use:"audit",Short:"Audit cluster health and configuration",Args:cobra.NoArgs,RunE:func(cmd *cobra.Command,_ []string)error{
		config,err:=fleetk8s.LoadKubeconfig(kubeconfigPath);if err!=nil{return err}
		contexts:=[]string{contextName}
		if all { contexts=sortedContextNames(config.Contexts) } else if contextName=="" { return errors.New("provide --context or use --all") }
		type result struct{findings []audit.Finding}
		results:=make(chan result,len(contexts));group,ctx:=errgroup.WithContext(cmd.Context());group.SetLimit(8)
		for _,name:=range contexts{name:=name;group.Go(func()error{
			client,clientErr:=fleetk8s.NewClient(kubeconfigPath,name)
			if clientErr!=nil{results<-result{[]audit.Finding{{Cluster:name,Severity:audit.SeverityHigh,Rule:"cluster-unreachable",Resource:name,Message:clientErr.Error()}}};return nil}
			requestCtx,cancel:=context.WithTimeout(ctx,30*time.Second);defer cancel()
			snapshot,snapshotErr:=fleetk8s.Snapshot(requestCtx,client,name)
			if snapshotErr!=nil{results<-result{[]audit.Finding{{Cluster:name,Severity:audit.SeverityHigh,Rule:"cluster-unreachable",Resource:name,Message:snapshotErr.Error()}}};return nil}
			minimum,configErr:=minimumKernelVersion(kubeconfigPath);if configErr!=nil{return configErr}
			results<-result{audit.Analyze(snapshot,minimum,time.Now())};return nil
		})}
		if err:=group.Wait();err!=nil{return err};close(results)
		findings:=[]audit.Finding{};for result:=range results{findings=append(findings,result.findings)}
		sort.Slice(findings,func(i,j int)bool{if findings[i].Cluster!=findings[j].Cluster{return findings[i].Cluster<findings[j].Cluster};if findings[i].Severity!=findings[j].Severity{return findings[i].Severity<findings[j].Severity};return findings[i].Rule<findings[j].Rule})
		tw:=tabwriter.NewWriter(cmd.OutOrStdout(),0,4,2,' ',0);fmt.Fprintln(tw,"CLUSTER\tSEVERITY\tRULE\tRESOURCE\tMESSAGE")
		high:=false;for _,finding:=range findings{fmt.Fprintf(tw,"%s\t%s\t%s\t%s\t%s\n",finding.Cluster,finding.Severity,finding.Rule,finding.Resource,finding.Message);if finding.Severity==audit.SeverityHigh{high=true}}
		if err:=tw.Flush();err!=nil{return err};if high{return ErrHighFindings};return nil
	}}
	command.Flags().BoolVar(&all,"all",false,"Audit all contexts in kubeconfig in parallel")
	command.Flags().StringVarP(&contextName,"context","c","","Kubeconfig context")
	return command
}

func newDrainCommand() *cobra.Command {
	var contextName string
	var yes bool
	var timeout time.Duration
	command:=&cobra.Command{Use:"drain NODE",Short:"Safely drain a node (dry-run by default)",Args:cobra.ExactArgs(1),RunE:func(cmd *cobra.Command,args []string)error{
		if contextName==""{return errors.New("--context is required")}
		client,err:=fleetk8s.NewClient(kubeconfigPath,contextName);if err!=nil{return err}
		ctx,cancel:=context.WithTimeout(cmd.Context(),timeout+30*time.Second);defer cancel()
		pods,err:=fleetk8s.DrainNode(ctx,client,args[0],!yes,timeout)
		for _,pod:=range pods{fmt.Fprintln(cmd.OutOrStdout(),pod)}
		if err!=nil{return err}
		if !yes{fmt.Fprintf(cmd.OutOrStdout(),"Dry run only. Pass --yes to evict pods.\nUncordon command: kubectl --context %s uncordon %s\n",contextName,args[0]);return nil}
		fmt.Fprintf(cmd.OutOrStdout(),"Drain complete. Node remains cordoned.\nUncordon command: kubectl --context %s uncordon %s\n",contextName,args[0]);return nil
	}}
	command.Flags().StringVarP(&contextName,"context","c","","Kubeconfig context")
	command.Flags().BoolVar(&yes,"yes",false,"Cordon and evict pods; without this flag only show a dry run")
	command.Flags().DurationVar(&timeout,"timeout",5*time.Minute,"Maximum time to wait for pods to leave")
	return command
}

func writeOutput(w io.Writer, format string, value any, table func(io.Writer)error)error{
	switch strings.ToLower(format){case "table","":return table(w);case "json":encoder:=json.NewEncoder(w);encoder.SetIndent("","  ");return encoder.Encode(value);default:return fmt.Errorf("unsupported output format %q (use table or json)",format)}
}

func sortedContextNames(contexts map[string]*clientcmdapi.Context)[]string{names:=make([]string,0,len(contexts));for name:=range contexts{names=append(names,name)};sort.Strings(names);return names}

func minor(version string)int{version=strings.TrimPrefix(version,"v");parts:=strings.Split(version,".");if len(parts)<2{return -1};var value int;_,err:=fmt.Sscanf(parts[1],"%d",&value);if err!=nil{return -1};return value}

func minimumKernelVersion(_ string) (string, error) {
	const configPath = "fleetctl.yaml"
	data, err := os.ReadFile(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return "5.15.0", nil
	}
	if err != nil {
		return "", err
	}
	var config struct {
		MinimumKernelVersion string `yaml:"minimumKernelVersion"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", err
	}
	if config.MinimumKernelVersion == "" { return "5.15.0", nil }
	return config.MinimumKernelVersion, nil
}




