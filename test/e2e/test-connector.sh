#!/bin/bash

# Test script for the HakonGo Kubernetes Connector
# This script helps verify that metrics are being collected and sent to the mock API

set -e

# Colors for better output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== HakonGo Kubernetes Connector Test Script ===${NC}"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if we can connect to the Kubernetes cluster
echo -e "\n${YELLOW}Checking Kubernetes connection...${NC}"
if ! kubectl get nodes &> /dev/null; then
    echo -e "${RED}Error: Cannot connect to Kubernetes cluster${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Connected to Kubernetes cluster${NC}"

# Deploy mock API if not already deployed
echo -e "\n${YELLOW}Checking mock API deployment...${NC}"
if ! kubectl get deployment mock-hakongo-api &> /dev/null; then
    echo -e "Mock API not found, deploying..."
    kubectl apply -f ../manifests/mock-api.yaml
    echo -e "Waiting for mock API to be ready..."
    kubectl wait --for=condition=available deployment/mock-hakongo-api --timeout=60s
else
    echo -e "Mock API already deployed, updating configuration..."
    kubectl apply -f ../manifests/mock-api.yaml
fi
echo -e "${GREEN}✓ Mock API is running${NC}"

# Get the mock API service details
MOCK_API_IP=$(kubectl get service mock-hakongo-api -o jsonpath='{.spec.clusterIP}')
MOCK_API_PORT=$(kubectl get service mock-hakongo-api -o jsonpath='{.spec.ports[?(@.name=="api")].port}')
MOCK_API_DASHBOARD_PORT=$(kubectl get service mock-hakongo-api -o jsonpath='{.spec.ports[?(@.name=="dashboard")].port}')

echo -e "Mock API is available at: http://${MOCK_API_IP}:${MOCK_API_PORT}"
echo -e "Mock API Dashboard is available at: http://${MOCK_API_IP}:${MOCK_API_DASHBOARD_PORT}"

# Check if connector is deployed
echo -e "\n${YELLOW}Checking connector deployment...${NC}"
if ! kubectl get deployment kubernetes-hakongo-connector-controller-manager -n kubernetes-hakongo-connector-system &> /dev/null 2>&1; then
    echo -e "${YELLOW}Connector not deployed, checking if we need to install CRDs...${NC}"
    
    # Check if CRDs are installed
    if ! kubectl get crd connectorconfigs.hakongo.com &> /dev/null; then
        echo -e "Installing connector CRDs..."
        kubectl apply -f ../manifests/hakongo.com_connectorconfigs.yaml
    fi
    
    echo -e "Deploying connector..."
    kubectl apply -f ../manifests/role.yaml
    kubectl apply -f ../manifests/manager.yaml
    
    # Create a secret for the API key
    kubectl create namespace kubernetes-hakongo-connector-system --dry-run=client -o yaml | kubectl apply -f -
    kubectl create secret generic hakongo-api-key \
        --from-literal=api-key=test-api-key \
        --namespace kubernetes-hakongo-connector-system \
        --dry-run=client -o yaml | kubectl apply -f -
    
    echo -e "Waiting for connector to be ready..."
    kubectl wait --for=condition=available deployment/kubernetes-hakongo-connector-controller-manager \
        --namespace kubernetes-hakongo-connector-system --timeout=120s
else
    echo -e "Connector already deployed"
fi
echo -e "${GREEN}✓ Connector is running${NC}"

# Update the connector config to use our mock API
echo -e "\n${YELLOW}Configuring connector to use mock API...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: hakongo.com/v1alpha1
kind: ConnectorConfig
metadata:
  name: test-connector-config
spec:
  apiEndpoint: http://${MOCK_API_IP}:${MOCK_API_PORT}
  apiKeySecret:
    name: hakongo-api-key
    namespace: kubernetes-hakongo-connector-system
  collectionInterval: 30
  enablePrometheusCollection: true
  prometheusEndpoint: http://prometheus-server.prometheus.svc.cluster.local:9090
EOF
echo -e "${GREEN}✓ Connector configured${NC}"

# Wait for metrics to be collected
echo -e "\n${YELLOW}Waiting for metrics collection (30 seconds)...${NC}"
sleep 30

# Check if metrics are being received by the mock API
echo -e "\n${YELLOW}Checking if metrics are being received...${NC}"
METRICS_COUNT=$(kubectl exec -it $(kubectl get pod -l app=mock-hakongo-api -o jsonpath='{.items[0].metadata.name}') \
    -c metrics-processor -- cat /var/lib/metrics/processed/count.txt 2>/dev/null || echo "0")

if [ "$METRICS_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Metrics are being received by the mock API${NC}"
    echo -e "Total metrics count: ${METRICS_COUNT}"
    
    # Get resource type counts
    echo -e "\n${YELLOW}Resource metrics breakdown:${NC}"
    kubectl exec -it $(kubectl get pod -l app=mock-hakongo-api -o jsonpath='{.items[0].metadata.name}') \
        -c metrics-processor -- sh -c "echo -e 'Pods: \t\t\t$(cat /var/lib/metrics/processed/pods.txt 2>/dev/null || echo 0)'"
    kubectl exec -it $(kubectl get pod -l app=mock-hakongo-api -o jsonpath='{.items[0].metadata.name}') \
        -c metrics-processor -- sh -c "echo -e 'Nodes: \t\t\t$(cat /var/lib/metrics/processed/nodes.txt 2>/dev/null || echo 0)'"
    kubectl exec -it $(kubectl get pod -l app=mock-hakongo-api -o jsonpath='{.items[0].metadata.name}') \
        -c metrics-processor -- sh -c "echo -e 'Services: \t\t$(cat /var/lib/metrics/processed/services.txt 2>/dev/null || echo 0)'"
    kubectl exec -it $(kubectl get pod -l app=mock-hakongo-api -o jsonpath='{.items[0].metadata.name}') \
        -c metrics-processor -- sh -c "echo -e 'Ingresses: \t\t$(cat /var/lib/metrics/processed/ingresses.txt 2>/dev/null || echo 0)'"
    kubectl exec -it $(kubectl get pod -l app=mock-hakongo-api -o jsonpath='{.items[0].metadata.name}') \
        -c metrics-processor -- sh -c "echo -e 'Namespaces: \t\t$(cat /var/lib/metrics/processed/namespaces.txt 2>/dev/null || echo 0)'"
    kubectl exec -it $(kubectl get pod -l app=mock-hakongo-api -o jsonpath='{.items[0].metadata.name}') \
        -c metrics-processor -- sh -c "echo -e 'PersistentVolumes: \t$(cat /var/lib/metrics/processed/pvs.txt 2>/dev/null || echo 0)'"
    kubectl exec -it $(kubectl get pod -l app=mock-hakongo-api -o jsonpath='{.items[0].metadata.name}') \
        -c metrics-processor -- sh -c "echo -e 'Workloads: \t\t$(cat /var/lib/metrics/processed/workloads.txt 2>/dev/null || echo 0)'"
else
    echo -e "${RED}✗ No metrics received by the mock API${NC}"
    echo -e "Please check the connector logs for errors:"
    echo -e "kubectl logs -n kubernetes-hakongo-connector-system deployment/kubernetes-hakongo-connector-controller-manager -c manager"
fi

# Provide instructions for accessing the dashboard
echo -e "\n${YELLOW}To access the metrics dashboard:${NC}"
echo -e "1. Run: kubectl port-forward service/mock-hakongo-api 8080:8080"
echo -e "2. Open in your browser: http://localhost:8080"
echo -e "3. The dashboard will show all collected metrics and their breakdown by resource type"

echo -e "\n${YELLOW}To view connector logs:${NC}"
echo -e "kubectl logs -n kubernetes-hakongo-connector-system deployment/kubernetes-hakongo-connector-controller-manager -c manager"

echo -e "\n${GREEN}Test script completed${NC}"
