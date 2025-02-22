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

	// Collectors specifies which resource collectors to enable
	// +optional
	Collectors CollectorConfig `json:"collectors,omitempty"`

	// HakonGo configuration
	// +kubebuilder:validation:Required
	HakonGo HakonGoConfig `json:"hakongo"`

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
}

// HakonGoConfig defines the configuration for connecting to HakonGo
type HakonGoConfig struct {
	// BaseURL is the base URL of the HakonGo API
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=^https?://.*
	BaseURL string `json:"baseUrl"`

	// ClusterID is the unique identifier for this cluster in HakonGo
	// +kubebuilder:validation:Required
	ClusterID string `json:"clusterId"`

	// APIKeySecret references the secret containing the API key
	// +kubebuilder:validation:Required
	APIKeySecret SecretKeyRef `json:"apiKeySecret"`
}

// SecretKeyRef references a key in a Secret
type SecretKeyRef struct {
	// Name is the name of the secret
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the namespace of the secret
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// Key is the key in the secret
	// +kubebuilder:validation:Required
	Key string `json:"key"`
}

// +kubebuilder:object:generate=true

// CollectorConfig specifies which collectors are enabled and their configurations
type CollectorConfig struct {
	// EnablePodMetrics enables collection of pod metrics
	// +kubebuilder:default=true
	EnablePodMetrics bool `json:"enablePodMetrics"`

	// EnableNodeMetrics enables collection of node metrics
	// +kubebuilder:default=true
	EnableNodeMetrics bool `json:"enableNodeMetrics"`

	// EnablePVMetrics enables collection of persistent volume metrics
	EnablePVMetrics bool `json:"enablePVMetrics"`

	// EnablePVCMetrics enables collection of PVC metrics
	// +kubebuilder:default=true
	EnablePVCMetrics bool `json:"enablePVCMetrics"`

	// EnableServiceMetrics enables collection of service metrics
	// +kubebuilder:default=true
	EnableServiceMetrics bool `json:"enableServiceMetrics"`
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
	// LastCollectionTime is the timestamp of the last successful metrics collection
	// +optional
	LastCollectionTime *metav1.Time `json:"lastCollectionTime,omitempty"`
	// MetricsCollected is the number of metrics collected
	// +optional
	MetricsCollected int32 `json:"metricsCollected,omitempty"`
	// LastError is the error message of the last collection attempt
	// +optional
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
