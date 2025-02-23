# Ingress Metrics API

Endpoint for submitting ingress metrics data.

## Endpoint

```
POST /api/v1/kubernetes/ingresses
```

## Request Format

```bash
curl -X POST \
  'http://localhost:3000/api/v1/kubernetes/ingresses' \
  -H 'X-API-Key: your-api-key' \
  -H 'Content-Type: application/json' \
  -d '{
    "cluster_id": "my-cluster",
    "collected_at": "2025-02-22T20:35:36Z",
    "resources": [
      {
        "metadata": {
          "name": "frontend-ingress",
          "namespace": "default",
          "uid": "ing-abc-123",
          "labels": {
            "app": "frontend",
            "environment": "production"
          }
        },
        "spec": {
          "ingress_class_name": "nginx",
          "rules": [
            {
              "host": "frontend.example.com",
              "http": {
                "paths": [
                  {
                    "path": "/",
                    "path_type": "Prefix",
                    "backend": {
                      "service": {
                        "name": "frontend-service",
                        "port": {
                          "number": 80
                        }
                      }
                    }
                  }
                ]
              }
            }
          ],
          "tls": [
            {
              "hosts": ["frontend.example.com"],
              "secret_name": "frontend-tls"
            }
          ]
        },
        "status": {
          "load_balancer": {
            "ingress": [
              {
                "hostname": "frontend.example.com",
                "ip": "203.0.113.1"
              }
            ]
          },
          "conditions": [
            {
              "type": "Ready",
              "status": "True",
              "last_transition_time": "2025-02-22T19:35:36Z"
            }
          ]
        },
        "metrics": {
          "traffic": {
            "requests": {
              "total": 15000,
              "rate": 25.5,
              "by_status": {
                "2xx": 14500,
                "3xx": 300,
                "4xx": 150,
                "5xx": 50
              }
            },
            "bytes": {
              "received": 1048576,
              "transmitted": 2097152,
              "rate": {
                "in": 10240,
                "out": 20480
              }
            },
            "latency": {
              "p50_ms": 50,
              "p90_ms": 100,
              "p99_ms": 200
            }
          },
          "ssl": {
            "handshakes": {
              "successful": 1000,
              "failed": 5,
              "reuses": 950
            },
            "certificate": {
              "expiry": "2026-02-22T20:35:36Z",
              "days_until_expiry": 365
            }
          },
          "backends": {
            "healthy": 3,
            "unhealthy": 0,
            "response_time_ms": 45
          }
        }
      }
    ]
  }'
```

## Field Descriptions

### Resource Metadata
- `name`: Name of the ingress
- `namespace`: Kubernetes namespace
- `uid`: Unique identifier
- `labels`: Key-value pairs of ingress labels

### Ingress Spec
- `ingress_class_name`: Name of the ingress class
- `rules`: Ingress routing rules
- `tls`: TLS configuration

### Ingress Status
- `load_balancer`: Load balancer status
- `conditions`: List of ingress conditions

### Metrics
- `traffic`:
  - `requests`:
    - `total`: Total number of requests
    - `rate`: Requests per second
    - `by_status`: Request counts by HTTP status code
  - `bytes`:
    - `received`: Total bytes received
    - `transmitted`: Total bytes transmitted
    - `rate`: Current transfer rates
  - `latency`: Request latency percentiles
- `ssl`:
  - `handshakes`: SSL handshake statistics
  - `certificate`: SSL certificate information
- `backends`:
  - `healthy`: Number of healthy backends
  - `unhealthy`: Number of unhealthy backends
  - `response_time_ms`: Average backend response time
