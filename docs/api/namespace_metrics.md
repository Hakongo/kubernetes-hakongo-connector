# Namespace Metrics API

Endpoint: `POST /api/v1/metrics/namespaces`

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
      "name": "string",
      "uid": "string",
      "status": {
        "phase": "string"
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
          "running": 0,
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
          },
          "storage": {
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
        "quotas": [
          {
            "name": "string",
            "resource": "string",
            "hard": 0,
            "used": 0
          }
        ]
      }
    }
  ]
}
```

## Resource Fields

### Namespace Resource

| Field | Type | Description |
|-------|------|-------------|
| name | string | Name of the namespace |
| uid | string | Unique identifier for the namespace |
| status | object | Current status of the namespace |
| metadata | object | Namespace metadata |
| metrics | object | Namespace metrics |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| phase | string | Current phase (Active, Terminating) |

### Metrics Fields

| Field | Type | Description |
|-------|------|-------------|
| pods.total | integer | Total number of pods |
| pods.running | integer | Number of running pods |
| pods.pending | integer | Number of pending pods |
| pods.failed | integer | Number of failed pods |
| pods.succeeded | integer | Number of succeeded pods |
| resources.cpu.usage | float | Current CPU usage in cores |
| resources.cpu.requests | float | Total CPU requests in cores |
| resources.cpu.limits | float | Total CPU limits in cores |
| resources.memory.usage_bytes | integer | Current memory usage in bytes |
| resources.memory.requests_bytes | integer | Total memory requests in bytes |
| resources.memory.limits_bytes | integer | Total memory limits in bytes |
| resources.storage.usage_bytes | integer | Current storage usage in bytes |
| resources.storage.requests_bytes | integer | Total storage requests in bytes |
| resources.storage.limits_bytes | integer | Total storage limits in bytes |
| network.rx_bytes | integer | Network bytes received |
| network.tx_bytes | integer | Network bytes transmitted |
| network.rx_packets | integer | Network packets received |
| network.tx_packets | integer | Network packets transmitted |
| network.rx_errors | integer | Receive errors count |
| network.tx_errors | integer | Transmit errors count |
| quotas | array | Resource quota usage |

### Quota Fields

| Field | Type | Description |
|-------|------|-------------|
| name | string | Name of the quota |
| resource | string | Resource type |
| hard | integer | Hard limit |
| used | integer | Current usage |

## Example Response

```json
{
  "status": "success",
  "message": "Metrics received successfully",
  "metrics_count": 1
}
