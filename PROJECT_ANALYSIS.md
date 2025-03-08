# Kubernetes HakonGo Connector - Project Analysis

## Project Structure

### Core Components

#### 1. API Definitions (`api/v1alpha1/`)
- **types.go**: Core CRD types
  - `ConnectorConfig`: Main configuration CRD
  - `ConnectorConfigSpec`: Configuration specification
  - `ConnectorConfigStatus`: Runtime status
- **groupversion_info.go**: API group version info
- **zz_generated.deepcopy.go**: Generated DeepCopy implementations

#### 2. Command Line Tools (`cmd/`)
- **controller/main.go**: Main controller entry point
  - Initializes manager and metrics server
  - Sets up controller reconciliation
  - Configures health probes

#### 3. Internal Packages (`internal/`)

##### 3.1 API Client (`internal/api/`)
- **client.go**: PRIMARY SaaS API client
  - Handles all API communication
  - Supports metrics submission
  - Retrieves cluster-specific configurations
  - Manages cost and alerting rules

##### 3.2 Cluster Management (`internal/cluster/`)
- **types.go**: Cluster context types
- **provider.go**: Cluster context provider
  - Gathers cluster information
  - Detects node groups
  - Extracts cloud metadata

##### 3.3 Metrics Collection (`internal/collector/`)
- **types.go**: Core metric types
- **Collectors**:
  - `pod_collector.go`: Pod metrics (uses metrics-server)
  - `node_collector.go`: Node metrics (uses metrics-server)
  - `pv_collector.go`: Persistent Volume metrics (uses metrics-server)
  - `service_collector.go`: Service metrics (uses metrics-server)
  - `namespace_collector.go`: Namespace metrics (basic info only)
  - `workload_collector.go`: Workload metrics (basic info only)
  - `ingress_collector.go`: Ingress metrics (basic info only)

##### 3.4 Controller (`internal/controller/`)
- **connector_controller.go**: Main controller logic
  - Manages ConnectorConfig resources
  - Coordinates metric collection
  - Handles API communication

## Testing Instructions

### Prerequisites
1. A running Kubernetes cluster (e.g., minikube, kind, or remote cluster)
2. kubectl configured to access the cluster
3. metrics-server installed in the cluster
4. HakonGo API credentials

### Local Development Setup
1. Start local Kubernetes cluster:
   ```bash
   minikube start --addons=metrics-server
   ```

2. Create API key secret:
   ```bash
   kubectl create secret generic hakongo-api-key \
     --from-literal=api-key=your-api-key \
     -n default
   ```

3. Apply CRD:
   ```bash
   make install
   ```

4. Run controller locally:
   ```bash
   make run
   ```

5. Apply sample configuration:
   ```bash
   kubectl apply -f config/samples/hakongo_v1alpha1_connectorconfig.yaml
   ```

### Testing
1. Unit Tests:
   ```bash
   make test
   ```

2. Integration Tests:
   ```bash
   make test-integration
   ```

3. E2E Tests (requires cluster):
   ```bash
   make test-e2e
   ```

### Deployment
1. Build container:
   ```bash
   make docker-build
   ```

2. Push to registry:
   ```bash
   make docker-push
   ```

3. Deploy to cluster:
   ```bash
   make deploy
   ```

## Recent Changes

### Fixed Issues
1. ✅ Removed duplicate API client implementation from `internal/hakongo/`
2. ✅ Generated CRD using controller-gen
3. ✅ Updated controller to use correct API client
4. ✅ Fixed collector initialization with correct client dependencies:
   - Pod, Node, PV, Service collectors: Use metrics-server client
   - Namespace, Workload, Ingress collectors: Basic Kubernetes client only

### Pending Tasks
1. [ ] Add more unit tests for collectors
2. [ ] Implement rate limiting for API calls
3. [ ] Add metrics for controller health
4. [ ] Add documentation for custom metrics

## Best Practices

1. **API Communication**
   - ALWAYS use `internal/api/client.go` for API communication
   - NEVER create new API client implementations

2. **Metric Collection**
   - All collectors MUST implement the `Collector` interface
   - Use common types from `collector/types.go`
   - Only use metrics-server client where resource metrics are needed

3. **Configuration**
   - All configuration through ConnectorConfig CRD
   - Secrets managed via Kubernetes secrets

4. **Error Handling**
   - Use detailed error wrapping
   - Update status conditions appropriately
   - Implement proper retries
