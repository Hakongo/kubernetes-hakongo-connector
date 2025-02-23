package cluster

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ContextProvider gathers cluster context information
type ContextProvider struct {
	kubeClient kubernetes.Interface
	config     *Config
}

// Config contains configuration for the context provider
type Config struct {
	// Name of the cluster (required)
	ClusterName string `json:"clusterName"`

	// Cloud provider name (e.g., aws, gcp, azure)
	ProviderName string `json:"providerName"`

	// Region where the cluster is running
	Region string `json:"region"`

	// Zone where the cluster is running (optional)
	Zone string `json:"zone,omitempty"`

	// Additional labels to apply to all metrics
	Labels map[string]string `json:"labels,omitempty"`

	// Additional metadata to include
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewContextProvider creates a new cluster context provider
func NewContextProvider(kubeClient kubernetes.Interface, config *Config) *ContextProvider {
	return &ContextProvider{
		kubeClient: kubeClient,
		config:     config,
	}
}

// GetContext gathers the current cluster context
func (p *ContextProvider) GetContext(ctx context.Context) (*ClusterContext, error) {
	// Get cluster version
	version, err := p.kubeClient.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	// Get node groups and their info
	nodes, err := p.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Map to track node groups
	nodeGroups := make(map[string]*NodeGroupInfo)

	// Process each node
	for _, node := range nodes.Items {
		// Get node group name from labels
		ngName := p.getNodeGroupName(node.Labels)
		if ngName == "" {
			continue
		}

		// Create or update node group info
		ng, exists := nodeGroups[ngName]
		if !exists {
			ng = &NodeGroupInfo{
				Name:   ngName,
				Labels: make(map[string]string),
			}
			nodeGroups[ngName] = ng
		}

		// Update platform info if not set
		if ng.Platform.OS == "" {
			ng.Platform = PlatformInfo{
				OS:           node.Status.NodeInfo.OperatingSystem,
				Architecture: node.Status.NodeInfo.Architecture,
				Version:      node.Status.NodeInfo.KubeletVersion,
			}
		}

		// Merge labels that are common across all nodes in the group
		if len(ng.Labels) == 0 {
			ng.Labels = node.Labels
		} else {
			ng.Labels = p.mergeCommonLabels(ng.Labels, node.Labels)
		}

		// Add provider-specific metadata
		ng.Metadata = p.getNodeGroupMetadata(node.Labels)
	}

	// Convert node groups map to slice
	ngSlice := make([]NodeGroupInfo, 0, len(nodeGroups))
	for _, ng := range nodeGroups {
		ngSlice = append(ngSlice, *ng)
	}

	// Build the complete cluster context
	return &ClusterContext{
		Name:             p.config.ClusterName,
		KubernetesVersion: version.GitVersion,
		Provider: ProviderInfo{
			Name:   p.config.ProviderName,
			Region: p.config.Region,
			Zone:   p.config.Zone,
		},
		Labels:     p.config.Labels,
		NodeGroups: ngSlice,
		Metadata:   p.config.Metadata,
	}, nil
}

// getNodeGroupName extracts the node group name from node labels
func (p *ContextProvider) getNodeGroupName(labels map[string]string) string {
	// Check common node group label patterns
	patterns := []string{
		"eks.amazonaws.com/nodegroup",    // EKS
		"cloud.google.com/gke-nodepool",  // GKE
		"agentpool",                      // AKS
		"node-pool",                      // Generic
	}

	for _, pattern := range patterns {
		if name, ok := labels[pattern]; ok {
			return name
		}
	}

	return ""
}

// mergeCommonLabels keeps only labels that are common between current and new labels
func (p *ContextProvider) mergeCommonLabels(current, new map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range current {
		if newV, ok := new[k]; ok && v == newV {
			result[k] = v
		}
	}
	return result
}

// getNodeGroupMetadata extracts provider-specific metadata from node labels
func (p *ContextProvider) getNodeGroupMetadata(labels map[string]string) map[string]interface{} {
	metadata := make(map[string]interface{})

	switch strings.ToLower(p.config.ProviderName) {
	case "aws":
		if instanceType, ok := labels["node.kubernetes.io/instance-type"]; ok {
			metadata["instance_type"] = instanceType
		}
		if zone, ok := labels["topology.kubernetes.io/zone"]; ok {
			metadata["availability_zone"] = zone
		}
	case "gcp":
		if machineType, ok := labels["beta.kubernetes.io/instance-type"]; ok {
			metadata["machine_type"] = machineType
		}
		if zone, ok := labels["topology.kubernetes.io/zone"]; ok {
			metadata["zone"] = zone
		}
	case "azure":
		if vmSize, ok := labels["node.kubernetes.io/instance-type"]; ok {
			metadata["vm_size"] = vmSize
		}
		if zone, ok := labels["topology.kubernetes.io/zone"]; ok {
			metadata["zone"] = zone
		}
	}

	return metadata
}
