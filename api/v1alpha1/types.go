package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ConnectorConfig is the Schema for the connectorconfigs API
type ConnectorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectorConfigSpec   `json:"spec,omitempty"`
	Status ConnectorConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConnectorConfigList contains a list of ConnectorConfig
type ConnectorConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConnectorConfig `json:"items"`
}

//+k8s:deepcopy-gen=true

// ConnectorConfigSpec defines the desired state of ConnectorConfig
type ConnectorConfigSpec struct {
	// HakonGo configuration for connecting to the API
	HakonGo HakonGoConfig `json:"hakongo"`

	// ClusterContext contains information about the cluster
	ClusterContext ClusterContextConfig `json:"clusterContext"`

	// Collectors is a list of collectors to enable
	// +optional
	Collectors []CollectorSpec `json:"collectors,omitempty"`

	// Cost configuration for cost calculations
	// +optional
	Cost *CostConfig `json:"cost,omitempty"`
}

//+k8s:deepcopy-gen=true

// ClusterContextConfig defines how to identify and label the cluster
type ClusterContextConfig struct {
	// Name of the cluster
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Type of the cluster (e.g. aws, gcp, azure)
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// Region where the cluster is running
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// Zone where the cluster is running
	// +optional
	Zone string `json:"zone,omitempty"`

	// Labels to be added to all metrics
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Additional metadata about the cluster
	// +optional
	Metadata map[string]string `json:"metadata,omitempty"`
}

//+k8s:deepcopy-gen=true

// CollectorSpec defines configuration for a specific collector
type CollectorSpec struct {
	// Name of the collector
	Name string `json:"name"`

	// Collection interval in seconds
	// +optional
	// +kubebuilder:default=60
	Interval int32 `json:"interval,omitempty"`

	// Labels to be added to metrics from this collector
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

//+k8s:deepcopy-gen=true

// HakonGoConfig defines configuration for connecting to HakonGo API
type HakonGoConfig struct {
	// BaseURL is the base URL for the HakonGo API
	BaseURL string `json:"baseUrl"`

	// APIKey reference to the API key secret
	APIKey SecretKeyRef `json:"apiKey"`
}

//+k8s:deepcopy-gen=true

// SecretKeyRef references a key in a secret
type SecretKeyRef struct {
	// Name of the secret
	Name string `json:"name"`

	// Key within the secret
	Key string `json:"key"`
}

//+k8s:deepcopy-gen=true

// CostConfig specifies the configuration for cost calculations
type CostConfig struct {
	// Currency is the currency used for cost calculations
	// +kubebuilder:default="USD"
	// +optional
	Currency string `json:"currency,omitempty"`

	// PriceBook is the name of the price book to use
	// +optional
	PriceBook string `json:"priceBook,omitempty"`

	// Labels to be added to cost metrics
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Additional metadata for cost calculations
	// +optional
	Metadata map[string]string `json:"metadata,omitempty"`
}

//+k8s:deepcopy-gen=true

// ConnectorConfigStatus defines the observed state of ConnectorConfig
type ConnectorConfigStatus struct {
	// LastCollectionTime is the last time metrics were collected
	LastCollectionTime *metav1.Time `json:"lastCollectionTime,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ConnectorConfig{}, &ConnectorConfigList{})
}
