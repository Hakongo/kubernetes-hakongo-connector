# Kubernetes HakonGo Connector

A powerful Kubernetes connector that provides deep insights into your cluster operations through a unified SaaS portal. This connector enables real-time monitoring, cost optimization, and intelligent service management across your Kubernetes infrastructure.

## 🚀 Key Features

- **Comprehensive Cluster Insights**
  - Real-time resource utilization metrics
  - Pod and node status monitoring
  - Kubernetes event collection and analysis
  - Network traffic analysis
  - Storage usage tracking

- **Cost Management**
  - Resource cost allocation
  - Usage-based billing insights
  - Cost optimization recommendations
  - Budget tracking and alerts

- **Service Intelligence**
  - Automatic service dependency mapping
  - Cross-namespace service relationship visualization
  - Service mesh integration
  - Traffic flow analysis

- **Security & Compliance**
  - RBAC analysis
  - Security policy compliance checking
  - Certificate management monitoring
  - Security context verification

- **Performance Analytics**
  - Historical performance trends
  - Anomaly detection
  - Scaling recommendations
  - Resource optimization suggestions

## 🏗️ Architecture

The connector is built using a modular architecture:

```
kubernetes-hakongo-connector/
├── cmd/
│   └── controller/          # Main controller entry point
├── pkg/
│   ├── apis/               # Custom Resource Definitions
│   │   ├── v1alpha1/      # API versions
│   │   └── config/        # CRD configurations
│   └── controllers/        # Kubernetes controllers/operators
│       ├── cluster/       # Cluster-wide controllers
│       ├── workload/      # Workload-specific controllers
│       └── cost/          # Cost management controllers
├── internal/
│   ├── collector/          # Resource collectors
│   │   ├── metrics/       # Metrics collectors
│   │   ├── resources/     # K8s resource collectors
│   │   └── costs/         # Cost data collectors
│   ├── analyzer/          # Data analysis components
│   ├── metrics/           # Metrics processing
│   └── api/               # SaaS API client
├── manifests/             # Kubernetes deployment manifests
└── docs/                  # Documentation
```

## 🔧 Prerequisites

- Go 1.24 or higher
- Access to a Kubernetes cluster (1.24+)
- kubectl configured with cluster access
- Docker for building containers
- Helm 3.x (optional, for chart deployment)

## 🚦 Quick Start

1. **Install Go 1.24+**
   ```bash
   # MacOS
   brew install go
   
   # Verify installation
   go version
   ```

2. **Clone and Build**
   ```bash
   git clone https://github.com/hakongo/kubernetes-hakongo-connector
   cd kubernetes-hakongo-connector
   make build
   ```

3. **Configure**
   ```bash
   # Copy and edit the configuration
   cp config/config.example.yaml config/config.yaml
   ```

4. **Deploy**
   ```bash
   # Using kubectl
   make deploy
   
   # Or using Helm
   helm install hakongo-connector ./charts/hakongo-connector
   ```

## 💻 Development

### Building

```bash
# Build binary
make build

# Run tests
make test

# Run specific tests
go test -v ./internal/collector -run TestEventCollector
go test -v ./internal/api -run TestClient_SendEventMetrics

# Build Docker image
make docker-build
```

### Running Locally

```bash
# Run controller locally
make run

# Run with specific configuration
make run CONFIG_FILE=./config/dev.yaml
```

### Deployment

```bash
# Deploy to cluster
make deploy

# Undeploy
make undeploy
```

## 🔍 Monitoring & Debugging

The connector exposes several endpoints:

- Metrics: `:8080/metrics`
- Health: `:8081/healthz`
- Ready: `:8081/readyz`

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details.

## 📄 License

MIT License - see the [LICENSE](LICENSE) file for details

## 🔗 Links

- [Documentation](docs/README.md)
- [API Reference](docs/api/README.md)
- [Event Metrics API](docs/api/event_metrics.md)
- [Event Metrics Testing](docs/testing/event_metrics_testing.md)
- [Architecture Guide](docs/architecture.md)
- [Troubleshooting Guide](docs/troubleshooting.md)
