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

	// CollectionInterval is the interval at which metrics are collected (in seconds)
	// +kubebuilder:validation:Minimum=30
	// +kubebuilder:default=300
	CollectionInterval int32 `json:"collectionInterval"`

	// IncludeNamespaces is a list of namespaces to include in metrics collection
	// If empty, all namespaces will be included except those in ExcludeNamespaces
	// +optional
	IncludeNamespaces []string `json:"includeNamespaces,omitempty"`

	// ExcludeNamespaces is a list of namespaces to exclude from metrics collection
	// +optional
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`

	// Collectors specifies which resource collectors to enable
	// +optional
	Collectors CollectorConfig `json:"collectors,omitempty"`

	// CostConfig specifies the configuration for cost calculations
	// +optional
	CostConfig CostConfig `json:"costConfig,omitempty"`
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

	// LastCollectionStatus indicates the status of the last collection attempt
	// +optional
	LastCollectionStatus string `json:"lastCollectionStatus,omitempty"`

	// Conditions represent the latest available observations of an object's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:object:generate=true

// ConnectorConfigList contains a list of ConnectorConfig
type ConnectorConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConnectorConfig `json:"items"`
}
