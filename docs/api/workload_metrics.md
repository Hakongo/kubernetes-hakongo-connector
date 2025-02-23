# Workload Metrics API

Endpoint: `POST /api/v1/metrics/workloads`

## Request Format

```json
{
  "cluster_id": "string",
  "context": {
    // See README.md for cluster context structure
  },
  "collected_at": "2025-02-23T10:00:00Z",
  "resources": [
    {
      "namespace": "string",
      "name": "string",
      "uid": "string",
      "kind": "string",
      "status": {
        "replicas": 0,
        "ready_replicas": 0,
        "updated_replicas": 0,
        "available_replicas": 0,
        "unavailable_replicas": 0,
        "conditions": [
          {
            "type": "string",
            "status": "string",
            "reason": "string",
            "message": "string",
            "last_transition_time": "string"
          }
        ]
      },
      "spec": {
        "replicas": 0,
        "selector": {
          "key": "value"
        },
        "strategy": {
          "type": "string",
          "rolling_update": {
            "max_surge": 0,
            "max_unavailable": 0
          }
        }
      },
      "metadata": {
        "labels": {
          "key": "value"
        },
        "annotations": {
          "key": "value"
        },
        "creation_timestamp": "string"
      },
      "metrics": {
        "pods": {
          "total": 0,
          "ready": 0,
          "not_ready": 0,
          "pending": 0,
          "failed": 0,
          "succeeded": 0
        },
        "resources": {
          "cpu": {
            "usage": 0.0,
            "requests": 0.0,
            "limits": 0.0
          },
          "memory": {
            "usage_bytes": 0,
            "requests_bytes": 0,
            "limits_bytes": 0
          }
        },
        "network": {
          "rx_bytes": 0,
          "tx_bytes": 0,
          "rx_packets": 0,
          "tx_packets": 0,
          "rx_errors": 0,
          "tx_errors": 0
        },
        "health": {
          "available_percentage": 0.0,
          "ready_percentage": 0.0
        }
      }
    }
  ]
}
```

## Resource Fields

### Workload Resource

| Field | Type | Description |
|-------|------|-------------|
| namespace | string | Namespace where the workload is running |
| name | string | Name of the workload |
| uid | string | Unique identifier for the workload |
| kind | string | Type of workload (Deployment, StatefulSet, DaemonSet) |
| status | object | Current status of the workload |
| spec | object | Workload specification details |
| metadata | object | Workload metadata |
| metrics | object | Workload metrics |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| replicas | integer | Total number of replicas |
| ready_replicas | integer | Number of ready replicas |
| updated_replicas | integer | Number of updated replicas |
| available_replicas | integer | Number of available replicas |
| unavailable_replicas | integer | Number of unavailable replicas |
| conditions | array | List of current workload conditions |

### Spec Fields

| Field | Type | Description |
|-------|------|-------------|
| replicas | integer | Desired number of replicas |
| selector | object | Pod selector labels |
| strategy | object | Deployment strategy configuration |

### Metrics Fields

| Field | Type | Description |
|-------|------|-------------|
| pods.total | integer | Total number of pods |
| pods.ready | integer | Number of ready pods |
| pods.not_ready | integer | Number of not ready pods |
| pods.pending | integer | Number of pending pods |
| pods.failed | integer | Number of failed pods |
| pods.succeeded | integer | Number of succeeded pods |
| resources.cpu.usage | float | Current CPU usage in cores |
| resources.cpu.requests | float | Total CPU requests in cores |
| resources.cpu.limits | float | Total CPU limits in cores |
| resources.memory.usage_bytes | integer | Current memory usage in bytes |
| resources.memory.requests_bytes | integer | Total memory requests in bytes |
| resources.memory.limits_bytes | integer | Total memory limits in bytes |
| network.rx_bytes | integer | Network bytes received |
| network.tx_bytes | integer | Network bytes transmitted |
| network.rx_packets | integer | Network packets received |
| network.tx_packets | integer | Network packets transmitted |
| network.rx_errors | integer | Receive errors count |
| network.tx_errors | integer | Transmit errors count |
| health.available_percentage | float | Percentage of available replicas |
| health.ready_percentage | float | Percentage of ready replicas |

## Example Response

```json
{
  "status": "success",
  "message": "Metrics received successfully",
  "metrics_count": 1
}
