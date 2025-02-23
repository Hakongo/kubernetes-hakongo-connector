# Node Metrics API

Endpoint: `POST /api/v1/metrics/nodes`

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
        "conditions": [
          {
            "type": "string",
            "status": "string",
            "reason": "string",
            "message": "string",
            "last_transition_time": "string"
          }
        ],
        "addresses": [
          {
            "type": "string",
            "address": "string"
          }
        ],
        "capacity": {
          "cpu": "string",
          "memory": "string",
          "pods": "string",
          "ephemeral-storage": "string"
        },
        "allocatable": {
          "cpu": "string",
          "memory": "string",
          "pods": "string",
          "ephemeral-storage": "string"
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
        "cpu": {
          "usage": 0.0,
          "capacity": 0.0,
          "allocatable": 0.0
        },
        "memory": {
          "usage_bytes": 0,
          "capacity_bytes": 0,
          "allocatable_bytes": 0
        },
        "network": {
          "rx_bytes": 0,
          "tx_bytes": 0,
          "rx_errors": 0,
          "tx_errors": 0
        },
        "filesystem": {
          "usage_bytes": 0,
          "capacity_bytes": 0,
          "available_bytes": 0,
          "inodes_used": 0,
          "inodes_free": 0
        }
      }
    }
  ]
}
```

## Resource Fields

### Node Resource

| Field | Type | Description |
|-------|------|-------------|
| name | string | Name of the node |
| uid | string | Unique identifier for the node |
| status | object | Current status of the node |
| metadata | object | Node metadata |
| metrics | object | Resource usage metrics |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| conditions | array | List of current node conditions |
| addresses | array | List of node addresses |
| capacity | object | Total node capacity |
| allocatable | object | Allocatable resources |

### Metrics Fields

| Field | Type | Description |
|-------|------|-------------|
| cpu.usage | float | Current CPU usage in cores |
| cpu.capacity | float | Total CPU capacity in cores |
| cpu.allocatable | float | Allocatable CPU in cores |
| memory.usage_bytes | integer | Current memory usage in bytes |
| memory.capacity_bytes | integer | Total memory capacity in bytes |
| memory.allocatable_bytes | integer | Allocatable memory in bytes |
| network.rx_bytes | integer | Network bytes received |
| network.tx_bytes | integer | Network bytes transmitted |
| network.rx_errors | integer | Receive errors count |
| network.tx_errors | integer | Transmit errors count |
| filesystem.usage_bytes | integer | Filesystem usage in bytes |
| filesystem.capacity_bytes | integer | Total filesystem capacity in bytes |
| filesystem.available_bytes | integer | Available filesystem space in bytes |
| filesystem.inodes_used | integer | Used inodes count |
| filesystem.inodes_free | integer | Free inodes count |

## Example Response

```json
{
  "status": "success",
  "message": "Metrics received successfully",
  "metrics_count": 1
}
