# Pod Metrics API

Endpoint for submitting pod metrics data.

## Endpoint

```
POST /api/v1/metrics/pods
```

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
      "status": {
        "phase": "string",
        "conditions": [
          {
            "type": "string",
            "status": "string",
            "reason": "string",
            "message": "string",
            "last_transition_time": "string"
          }
        ],
        "container_statuses": [
          {
            "name": "string",
            "ready": true,
            "restarts": 0,
            "state": {
              "running": {
                "started_at": "string"
              }
            },
            "last_state": {}
          }
        ]
      },
      "spec": {
        "node_name": "string",
        "priority_class": "string",
        "service_account": "string"
      },
      "metadata": {
        "labels": {
          "key": "value"
        },
        "annotations": {
          "key": "value"
        },
        "owner_references": [
          {
            "kind": "string",
            "name": "string",
            "uid": "string"
          }
        ],
        "creation_timestamp": "string"
      },
      "metrics": {
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
        "network": {
          "rx_bytes": 0,
          "tx_bytes": 0
        },
        "filesystem": {
          "usage_bytes": 0,
          "capacity_bytes": 0
        }
      }
    }
  ]
}
```

## Resource Fields

### Pod Resource

| Field | Type | Description |
|-------|------|-------------|
| namespace | string | Namespace where the pod is running |
| name | string | Name of the pod |
| uid | string | Unique identifier for the pod |
| status | object | Current status of the pod |
| spec | object | Pod specification details |
| metadata | object | Pod metadata |
| metrics | object | Resource usage metrics |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| phase | string | Current phase of the pod (Pending, Running, Succeeded, Failed, Unknown) |
| conditions | array | List of current pod conditions |
| container_statuses | array | Status of each container in the pod |

### Metrics Fields

| Field | Type | Description |
|-------|------|-------------|
| cpu.usage | float | Current CPU usage in cores |
| cpu.requests | float | CPU requests in cores |
| cpu.limits | float | CPU limits in cores |
| memory.usage_bytes | integer | Current memory usage in bytes |
| memory.requests_bytes | integer | Memory requests in bytes |
| memory.limits_bytes | integer | Memory limits in bytes |
| network.rx_bytes | integer | Network bytes received |
| network.tx_bytes | integer | Network bytes transmitted |
| filesystem.usage_bytes | integer | Filesystem usage in bytes |
| filesystem.capacity_bytes | integer | Filesystem capacity in bytes |

## Example Response

```json
{
  "status": "success",
  "message": "Metrics received successfully",
  "metrics_count": 1
}
