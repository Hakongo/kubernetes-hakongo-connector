package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ConnectorConfig is the Schema for the connectorconfigs API
type ConnectorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectorConfigSpec   `json:"spec,omitempty"`
	Status ConnectorConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=true

// ConnectorConfigSpec defines the desired state of ConnectorConfig
type ConnectorConfigSpec struct {
	// SaaSEndpoint is the endpoint URL for the HakonGo SaaS platform
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^https?://.*`
	SaaSEndpoint string `json:"saasEndpoint"`

	// APIKey is the authentication key for the HakonGo SaaS platform
	// +kubebuilder:validation:Required
	APIKey string `json:"apiKey"`

	// HakonGo configuration
	// +kubebuilder:validation:Required
	HakonGo HakonGoConfig `json:"hakongo"`

	// Collectors specifies which collectors are enabled and their configurations
	Collectors []CollectorSpec `json:"collectors"`

	// IncludeNamespaces is a list of namespaces to include in metrics collection
	// If empty, all namespaces will be included except those in ExcludeNamespaces
	// +optional
	IncludeNamespaces []string `json:"includeNamespaces,omitempty"`

	// ExcludeNamespaces is a list of namespaces to exclude from metrics collection
	// +optional
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`

	// CollectionInterval is the interval at which metrics are collected (in seconds)
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=^[0-9]+[mh]$
	// +kubebuilder:default="5m"
	CollectionInterval string `json:"collectionInterval"`

	// CostConfig specifies the configuration for cost calculations
	// +optional
	CostConfig CostConfig `json:"costConfig,omitempty"`

	// Cluster context configuration
	ClusterContext ClusterContextConfig `json:"clusterContext"`
}

// ClusterContextConfig defines how to identify and label the cluster
type ClusterContextConfig struct {
	// Name of the cluster
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Cloud provider name (e.g., aws, gcp, azure)
	// +kubebuilder:validation:Required
	ProviderName string `json:"providerName"`

	// Region where the cluster is running
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// Zone where the cluster is running (optional)
	// +optional
	Zone string `json:"zone,omitempty"`

	// Additional labels to apply to all metrics
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Additional metadata to include with metrics
	// +optional
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CollectorSpec defines configuration for a specific collector
type CollectorSpec struct {
	// Name of the collector
	Name string `json:"name"`

	// Whether this collector is enabled
	Enabled bool `json:"enabled"`

	// Namespaces to include in collection (empty means all)
	IncludeNamespaces []string `json:"includeNamespaces,omitempty"`

	// Namespaces to exclude from collection
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`
}

// HakonGoConfig defines configuration for connecting to HakonGo API
type HakonGoConfig struct {
	// BaseURL is the base URL for the HakonGo API
	BaseURL string `json:"baseUrl"`

	// ClusterID is the unique identifier for this cluster
	ClusterID string `json:"clusterId"`

	// APIKeySecret references a secret containing the API key
	APIKeySecret SecretKeyRef `json:"apiKeySecret"`
}

// SecretKeyRef references a key in a secret
type SecretKeyRef struct {
	// Name of the secret
	Name string `json:"name"`

	// Key in the secret
	Key string `json:"key"`
}

// +kubebuilder:object:generate=true

// CostConfig specifies the configuration for cost calculations
type CostConfig struct {
	// Currency is the currency used for cost calculations
	// +kubebuilder:default="USD"
	Currency string `json:"currency"`

	// CPUCostPerCore is the cost per CPU core per hour
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0.04
	CPUCostPerCore float64 `json:"cpuCostPerCore"`

	// MemoryCostPerGB is the cost per GB of memory per hour
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0.01
	MemoryCostPerGB float64 `json:"memoryCostPerGB"`

	// StorageCostPerGB is the cost per GB of storage per hour
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0.0002
	StorageCostPerGB float64 `json:"storageCostPerGB"`
}

// +kubebuilder:object:generate=true

// ConnectorConfigStatus defines the observed state of ConnectorConfig
type ConnectorConfigStatus struct {
	// LastCollectionTime is the last time metrics were collected
	LastCollectionTime *metav1.Time `json:"lastCollectionTime,omitempty"`

	// MetricsCollected is the number of metrics collected in the last run
	MetricsCollected int64 `json:"metricsCollected,omitempty"`

	// LastError is the last error encountered during collection
	LastError string `json:"lastError,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:object:generate=true

// ConnectorConfigList contains a list of ConnectorConfig
type ConnectorConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConnectorConfig `json:"items"`
}
