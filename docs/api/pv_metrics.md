# Persistent Volume Metrics API

Endpoint: `POST /api/v1/metrics/pvs`

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
        "phase": "string",
        "reason": "string",
        "message": "string"
      },
      "spec": {
        "capacity": {
          "storage": "string"
        },
        "access_modes": ["string"],
        "storage_class_name": "string",
        "volume_mode": "string",
        "persistent_volume_reclaim_policy": "string"
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
        "storage": {
          "capacity_bytes": 0,
          "used_bytes": 0,
          "available_bytes": 0,
          "inodes_total": 0,
          "inodes_used": 0,
          "inodes_free": 0
        },
        "status": {
          "bound": true,
          "claim_namespace": "string",
          "claim_name": "string"
        }
      }
    }
  ]
}
```

## Resource Fields

### Persistent Volume Resource

| Field | Type | Description |
|-------|------|-------------|
| name | string | Name of the persistent volume |
| uid | string | Unique identifier for the volume |
| status | object | Current status of the volume |
| spec | object | Volume specification details |
| metadata | object | Volume metadata |
| metrics | object | Storage metrics |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| phase | string | Current phase (Available, Bound, Released, Failed) |
| reason | string | Reason for current phase |
| message | string | Human-readable message about current phase |

### Spec Fields

| Field | Type | Description |
|-------|------|-------------|
| capacity.storage | string | Storage capacity |
| access_modes | array | List of access modes |
| storage_class_name | string | Storage class name |
| volume_mode | string | Volume mode (Block, Filesystem) |
| persistent_volume_reclaim_policy | string | Reclaim policy |

### Metrics Fields

| Field | Type | Description |
|-------|------|-------------|
| storage.capacity_bytes | integer | Total storage capacity in bytes |
| storage.used_bytes | integer | Used storage in bytes |
| storage.available_bytes | integer | Available storage in bytes |
| storage.inodes_total | integer | Total number of inodes |
| storage.inodes_used | integer | Used inodes count |
| storage.inodes_free | integer | Free inodes count |
| status.bound | boolean | Whether the volume is bound |
| status.claim_namespace | string | Namespace of the bound claim |
| status.claim_name | string | Name of the bound claim |

## Example Response

```json
{
  "status": "success",
  "message": "Metrics received successfully",
  "metrics_count": 1
}
