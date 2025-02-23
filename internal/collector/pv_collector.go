package collector

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type PVCollector struct {
	kubeClient    kubernetes.Interface
	metricsClient versioned.Interface
	config        CollectorConfig
}

func NewPVCollector(kubeClient kubernetes.Interface, metricsClient versioned.Interface, config CollectorConfig) *PVCollector {
	return &PVCollector{
		kubeClient:    kubeClient,
		metricsClient: metricsClient,
		config:        config,
	}
}

func (pc *PVCollector) Name() string { return "pv-collector" }

func (pc *PVCollector) Description() string {
	return "Collects metrics for Kubernetes PersistentVolumes"
}

func (pc *PVCollector) Collect(ctx context.Context) ([]ResourceMetrics, error) {
	var metrics []ResourceMetrics

	pvs, err := pc.kubeClient.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list persistent volumes: %w", err)
	}

	pvcs, err := pc.kubeClient.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list persistent volume claims: %w", err)
	}

	// Create a map of PVC name to PVC for quick lookups
	pvcMap := make(map[string]*corev1.PersistentVolumeClaim)
	for i := range pvcs.Items {
		pvc := &pvcs.Items[i]
		key := fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)
		pvcMap[key] = pvc
	}

	for _, pv := range pvs.Items {
		metric := ResourceMetrics{
			Name:        pv.Name,
			Kind:        "PersistentVolume",
			Labels:      pv.Labels,
			CollectedAt: time.Now(),
		}

		// Calculate storage metrics
		metric.Storage = pc.calculateStorageMetrics(&pv)

		// Calculate cost metrics based on storage class and capacity
		metric.Cost = pc.calculateCostMetrics(&pv)

		// If PV is bound to a PVC, include PVC details
		if pv.Spec.ClaimRef != nil {
			pvcKey := fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
			if pvc, exists := pvcMap[pvcKey]; exists {
				metric.Storage.PVCName = pvc.Name
				metric.Namespace = pvc.Namespace
			}
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (pc *PVCollector) calculateStorageMetrics(pv *corev1.PersistentVolume) StorageMetrics {
	storage := StorageMetrics{
		CapacityBytes: pv.Spec.Capacity.Storage().Value(),
	}

	// Set usage based on phase
	switch pv.Status.Phase {
	case corev1.VolumeBound:
		storage.UsageBytes = storage.CapacityBytes // Assume full usage when bound
	case corev1.VolumeAvailable:
		storage.Available = storage.CapacityBytes
	}

	return storage
}

func (pc *PVCollector) calculateCostMetrics(pv *corev1.PersistentVolume) CostMetrics {
	// Base storage cost calculation
	storageGB := float64(pv.Spec.Capacity.Storage().Value()) / float64(1<<30)
	var costPerGBHour float64

	// Adjust cost based on storage class
	switch {
	case pv.Spec.StorageClassName == "premium-ssd":
		costPerGBHour = 0.17 // Premium SSD cost
	case pv.Spec.StorageClassName == "standard-ssd":
		costPerGBHour = 0.08 // Standard SSD cost
	default:
		costPerGBHour = 0.04 // Standard HDD cost
	}

	storageCost := storageGB * costPerGBHour

	// Add premium for block volumes
	if pv.Spec.VolumeMode != nil && *pv.Spec.VolumeMode == corev1.PersistentVolumeBlock {
		storageCost *= 1.2 // 20% premium for block storage
	}

	return CostMetrics{
		Currency:    "USD",
		StorageCost: storageCost,
		TotalCost:   storageCost,
	}
}
