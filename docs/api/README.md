# HakonGo Kubernetes Connector API

This document describes the API endpoints and data structures used by the HakonGo Kubernetes Connector.

## Common Request Structure

All metric requests to the HakonGo API share a common structure:

```json
{
  "cluster_id": "string",
  "context": {
    "name": "string",
    "kubernetes_version": "string",
    "provider": {
      "name": "string",
      "region": "string",
      "zone": "string"
    },
    "labels": {
      "key": "value"
    },
    "node_groups": [
      {
        "name": "string",
        "labels": {
          "key": "value"
        },
        "platform": {
          "os": "string",
          "architecture": "string",
          "version": "string"
        },
        "metadata": {
          "key": "value"
        }
      }
    ],
    "metadata": {
      "key": "value"
    }
  },
  "collected_at": "2025-02-23T10:00:00Z",
  "resources": []
}
```

### Common Fields

| Field | Type | Description |
|-------|------|-------------|
| cluster_id | string | Unique identifier for the cluster |
| context | object | Cluster context information |
| context.name | string | Name of the cluster |
| context.kubernetes_version | string | Version of Kubernetes running in the cluster |
| context.provider | object | Cloud provider information |
| context.provider.name | string | Name of the cloud provider (e.g., aws, gcp, azure) |
| context.provider.region | string | Region where the cluster is running |
| context.provider.zone | string | Zone where the cluster is running (optional) |
| context.labels | object | Key-value pairs of labels applied to the cluster |
| context.node_groups | array | List of node groups in the cluster |
| context.node_groups[].name | string | Name of the node group |
| context.node_groups[].labels | object | Labels applied to the node group |
| context.node_groups[].platform | object | Platform information for the node group |
| context.node_groups[].platform.os | string | Operating system of the nodes |
| context.node_groups[].platform.architecture | string | CPU architecture of the nodes |
| context.node_groups[].platform.version | string | Kubernetes version of the nodes |
| context.node_groups[].metadata | object | Provider-specific metadata for the node group |
| context.metadata | object | Additional cluster metadata |
| collected_at | string | ISO 8601 timestamp when metrics were collected |
| resources | array | Array of resource metrics (see specific collector docs) |

## Authentication

All requests must include an authentication token in the Authorization header:

```
Authorization: Bearer <api-key>
```

## API Endpoints

Each collector has its own endpoint and specific resource format. See the following documents for details:

- [Pod Metrics](pod_metrics.md)
- [Node Metrics](node_metrics.md)
- [Persistent Volume Metrics](pv_metrics.md)
- [Service Metrics](service_metrics.md)
- [Namespace Metrics](namespace_metrics.md)
- [Workload Metrics](workload_metrics.md)
- [Ingress Metrics](ingress_metrics.md)

## Error Responses

The API uses standard HTTP status codes and returns errors in the following format:

```json
{
  "error": {
    "code": "string",
    "message": "string"
  }
}
```

### Common Error Codes

| Status Code | Description |
|------------|-------------|
| 400 | Bad Request - Invalid request format or missing required fields |
| 401 | Unauthorized - Invalid or missing API key |
| 403 | Forbidden - Valid API key but insufficient permissions |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server-side error |
