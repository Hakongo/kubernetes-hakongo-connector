package collector

import (
	"context"
	"time"
)

// ResourceMetrics represents collected metrics for a Kubernetes resource
type ResourceMetrics struct {
	// Resource metadata
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Kind      string            `json:"kind"`
	Labels    map[string]string `json:"labels"`

	// Collection metadata
	CollectedAt time.Time `json:"collected_at"`

	// Resource usage metrics
	CPU     CPUMetrics     `json:"cpu,omitempty"`
	Memory  MemoryMetrics  `json:"memory,omitempty"`
	Storage StorageMetrics `json:"storage,omitempty"`
	Network NetworkMetrics `json:"network,omitempty"`

	// Cost related information
	Cost CostMetrics `json:"cost,omitempty"`

	// Container metrics
	Containers []ContainerMetrics `json:"containers,omitempty"`

	// Status information
	Status map[string]interface{} `json:"status,omitempty"`
}

// CPUMetrics represents CPU usage metrics
type CPUMetrics struct {
	UsageNanoCores    int64   `json:"usageNanoCores"`
	UsageCorePercent  float64 `json:"usageCorePercent"`
	RequestMilliCores int64   `json:"requestMilliCores"`
	LimitMilliCores   int64   `json:"limitMilliCores"`
	ThrottlingSeconds float64 `json:"throttlingSeconds"`
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	UsageBytes      int64 `json:"usageBytes"`
	RequestBytes    int64 `json:"requestBytes"`
	LimitBytes      int64 `json:"limitBytes"`
	RSSBytes        int64 `json:"rssBytes"`
	PageFaults      int64 `json:"pageFaults"`
	MajorPageFaults int64 `json:"majorPageFaults"`
}

// StorageMetrics represents storage usage metrics
type StorageMetrics struct {
	UsageBytes    int64  `json:"usageBytes"`
	CapacityBytes int64  `json:"capacityBytes"`
	Available     int64  `json:"available"`
	PVCName       string `json:"pvcName,omitempty"`
	DiskPressure  bool   `json:"diskPressure"`
}

// NetworkMetrics represents network usage metrics
type NetworkMetrics struct {
	RxBytes   int64 `json:"rxBytes"`
	TxBytes   int64 `json:"txBytes"`
	RxPackets int64 `json:"rxPackets"`
	TxPackets int64 `json:"txPackets"`
	RxErrors  int64 `json:"rxErrors"`
	TxErrors  int64 `json:"txErrors"`
	RxDropped int64 `json:"rxDropped"`
	TxDropped int64 `json:"txDropped"`
}

// CostMetrics represents cost-related metrics
type CostMetrics struct {
	CPUCost     float64 `json:"cpuCost"`
	MemoryCost  float64 `json:"memoryCost"`
	StorageCost float64 `json:"storageCost"`
	NetworkCost float64 `json:"networkCost"`
	TotalCost   float64 `json:"totalCost"`
	Currency    string  `json:"currency"`
}

// ContainerMetrics represents container metrics
type ContainerMetrics struct {
	Name     string        `json:"name"`
	CPU      CPUMetrics    `json:"cpu"`
	Memory   MemoryMetrics `json:"memory"`
	Ready    bool          `json:"ready"`
	Restarts int32         `json:"restarts"`
	State    string        `json:"state"`
}

// Collector interface defines methods that must be implemented by resource collectors
type Collector interface {
	// Collect gathers metrics for the specified resource
	Collect(ctx context.Context) ([]ResourceMetrics, error)

	// Name returns the collector's name
	Name() string

	// Description returns the collector's description
	Description() string
}

// CollectorConfig represents common configuration for collectors
type CollectorConfig struct {
	// CollectionInterval is how often metrics should be collected
	CollectionInterval time.Duration

	// Namespaces to include in collection (empty means all)
	IncludeNamespaces []string `json:"include_namespaces"`

	// Namespaces to exclude from collection
	ExcludeNamespaces []string `json:"exclude_namespaces"`

	// Labels to include in collection (empty means all)
	IncludeLabels map[string]string

	// Resource types to collect metrics for
	ResourceTypes []string

	// MaxConcurrentCollections limits concurrent metric collection
	MaxConcurrentCollections int
}
