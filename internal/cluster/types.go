package cluster

// NodeGroupInfo contains information about a specific node group
type NodeGroupInfo struct {
	// Name of the node group
	Name string `json:"name"`

	// Labels specific to this node group
	Labels map[string]string `json:"labels"`

	// Platform information for this node group
	Platform PlatformInfo `json:"platform"`

	// Provider-specific metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PlatformInfo contains platform-specific information
type PlatformInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Version      string `json:"version,omitempty"`
}

// ProviderInfo contains cloud provider information
type ProviderInfo struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Zone   string `json:"zone,omitempty"`

	// Provider-specific metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ClusterContext contains all cluster-specific information
type ClusterContext struct {
	// Name of the cluster
	Name string `json:"name"`

	// Kubernetes version
	KubernetesVersion string `json:"kubernetes_version"`

	// Cloud provider information
	Provider ProviderInfo `json:"provider"`

	// Global cluster labels
	Labels map[string]string `json:"labels"`

	// Information about node groups
	NodeGroups []NodeGroupInfo `json:"node_groups"`

	// Additional cluster-wide metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
