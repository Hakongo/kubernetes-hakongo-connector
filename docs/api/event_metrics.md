# Event Metrics API

Endpoint: `POST /api/v1/metrics/events`

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
      "type": "string",
      "reason": "string",
      "message": "string",
      "source": {
        "component": "string",
        "host": "string"
      },
      "involved_object": {
        "kind": "string",
        "namespace": "string",
        "name": "string",
        "uid": "string"
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
        "count": 0,
        "first_timestamp": "string",
        "last_timestamp": "string",
        "duration_seconds": 0,
        "severity": "string"
      }
    }
  ]
}
```

## Resource Fields

### Event Resource

| Field | Type | Description |
|-------|------|-------------|
| namespace | string | Namespace where the event occurred |
| name | string | Name of the event |
| uid | string | Unique identifier for the event |
| type | string | Event type (Normal, Warning) |
| reason | string | Short reason for the event |
| message | string | Detailed event message |
| source | object | Source of the event |
| involved_object | object | Object that the event is about |
| metadata | object | Event metadata |
| metrics | object | Event metrics |

### Source Fields

| Field | Type | Description |
|-------|------|-------------|
| component | string | Component that generated the event |
| host | string | Host where the event occurred |

### Involved Object Fields

| Field | Type | Description |
|-------|------|-------------|
| kind | string | Kind of the involved object |
| namespace | string | Namespace of the involved object |
| name | string | Name of the involved object |
| uid | string | UID of the involved object |

### Metrics Fields

| Field | Type | Description |
|-------|------|-------------|
| count | integer | Number of times this event occurred |
| first_timestamp | string | Time when the event was first recorded |
| last_timestamp | string | Time when the event was last recorded |
| duration_seconds | integer | Duration between first and last occurrence |
| severity | string | Event severity level |

## Example Response

```json
{
  "status": "success",
  "message": "Metrics received successfully",
  "metrics_count": 1
}
```
