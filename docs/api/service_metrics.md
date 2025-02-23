# Service Metrics API

Endpoint: `POST /api/v1/metrics/services`

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
        "load_balancer": {
          "ingress": [
            {
              "ip": "string",
              "hostname": "string"
            }
          ]
        }
      },
      "spec": {
        "type": "string",
        "cluster_ip": "string",
        "external_ips": ["string"],
        "ports": [
          {
            "name": "string",
            "protocol": "string",
            "port": 0,
            "target_port": 0,
            "node_port": 0
          }
        ],
        "selector": {
          "key": "value"
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
        "endpoints": {
          "total": 0,
          "ready": 0,
          "not_ready": 0
        },
        "network": {
          "rx_bytes": 0,
          "tx_bytes": 0,
          "rx_packets": 0,
          "tx_packets": 0,
          "rx_errors": 0,
          "tx_errors": 0
        },
        "connections": {
          "active": 0,
          "total": 0
        }
      }
    }
  ]
}
```

## Resource Fields

### Service Resource

| Field | Type | Description |
|-------|------|-------------|
| namespace | string | Namespace where the service is running |
| name | string | Name of the service |
| uid | string | Unique identifier for the service |
| status | object | Current status of the service |
| spec | object | Service specification details |
| metadata | object | Service metadata |
| metrics | object | Service metrics |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| load_balancer.ingress | array | Load balancer ingress points |
| load_balancer.ingress[].ip | string | IP address of the ingress point |
| load_balancer.ingress[].hostname | string | Hostname of the ingress point |

### Spec Fields

| Field | Type | Description |
|-------|------|-------------|
| type | string | Service type (ClusterIP, NodePort, LoadBalancer) |
| cluster_ip | string | Cluster-internal IP |
| external_ips | array | List of external IPs |
| ports | array | List of exposed ports |
| selector | object | Pod selector labels |

### Metrics Fields

| Field | Type | Description |
|-------|------|-------------|
| endpoints.total | integer | Total number of endpoints |
| endpoints.ready | integer | Number of ready endpoints |
| endpoints.not_ready | integer | Number of not ready endpoints |
| network.rx_bytes | integer | Network bytes received |
| network.tx_bytes | integer | Network bytes transmitted |
| network.rx_packets | integer | Network packets received |
| network.tx_packets | integer | Network packets transmitted |
| network.rx_errors | integer | Receive errors count |
| network.tx_errors | integer | Transmit errors count |
| connections.active | integer | Number of active connections |
| connections.total | integer | Total connections since start |

## Example Response

```json
{
  "status": "success",
  "message": "Metrics received successfully",
  "metrics_count": 1
}
