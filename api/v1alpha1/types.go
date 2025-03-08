// +groupName=hakongo.com

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

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

	// Prometheus defines the configuration for Prometheus metrics
	// +optional
	Prometheus *PrometheusConfig `json:"prometheus,omitempty"`

	// MetricsServer defines the configuration for Kubernetes Metrics Server
	// +optional
	MetricsServer *MetricsServerConfig `json:"metricsServer,omitempty"`
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

// HakonGoConfig defines configuration for the HakonGo API
type HakonGoConfig struct {
	// BaseURL is the base URL for the HakonGo API
	// +kubebuilder:validation:Required
	BaseURL string `json:"baseURL"`

	// APIKey defines the API key for authentication
	// +kubebuilder:validation:Required
	APIKey corev1.SecretKeySelector `json:"apiKey"`
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

// PrometheusConfig defines configuration for Prometheus metrics collection
// +kubebuilder:object:generate=true
type PrometheusConfig struct {
	// URL is the base URL for the Prometheus API
	// +kubebuilder:validation:Required
	URL string `json:"url"`

	// ServiceMonitorSelector defines labels to select ServiceMonitors
	// +optional
	ServiceMonitorSelector map[string]string `json:"serviceMonitorSelector,omitempty"`

	// ScrapeInterval defines how frequently to scrape targets
	// +optional
	// +kubebuilder:default="30s"
	ScrapeInterval string `json:"scrapeInterval,omitempty"`

	// QueryTimeout defines the timeout for Prometheus queries
	// +optional
	// +kubebuilder:default="30s"
	QueryTimeout string `json:"queryTimeout,omitempty"`

	// BasicAuth defines basic authentication configuration
	// +optional
	BasicAuth *BasicAuthConfig `json:"basicAuth,omitempty"`

	// BearerToken defines bearer token authentication
	// +optional
	BearerToken *corev1.SecretKeySelector `json:"bearerToken,omitempty"`

	// TLSConfig defines TLS configuration
	// +optional
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
}

//+k8s:deepcopy-gen:interfaces=github.com/openshift/hive/pkg/apis/hive/v1alpha1.SecretKeySelector

// BasicAuthConfig defines basic authentication configuration
// +kubebuilder:object:generate=true
type BasicAuthConfig struct {
	// Username for basic authentication
	// +optional
	Username *corev1.SecretKeySelector `json:"username,omitempty"`

	// Password for basic authentication
	// +optional
	Password *corev1.SecretKeySelector `json:"password,omitempty"`
}

//+k8s:deepcopy-gen:interfaces=github.com/openshift/hive/pkg/apis/hive/v1alpha1.SecretKeySelector

// TLSConfig defines TLS configuration
// +kubebuilder:object:generate=true
type TLSConfig struct {
	// CA defines the CA certificate
	// +optional
	CA *corev1.SecretKeySelector `json:"ca,omitempty"`

	// Cert defines the client certificate
	// +optional
	Cert *corev1.SecretKeySelector `json:"cert,omitempty"`

	// Key defines the client key
	// +optional
	Key *corev1.SecretKeySelector `json:"key,omitempty"`

	// InsecureSkipVerify defines whether to skip TLS verification
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

//+k8s:deepcopy-gen=true

// MetricsServerConfig defines configuration for Kubernetes Metrics Server
// +kubebuilder:object:generate=true
type MetricsServerConfig struct {
	// Enabled defines whether to use Kubernetes Metrics Server
	// +optional
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`
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
