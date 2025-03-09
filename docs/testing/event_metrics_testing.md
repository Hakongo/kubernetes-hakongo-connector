# Testing Event Metrics Collection

This document describes how to test the event metrics collection functionality in the HakonGo Kubernetes connector.

## Overview

The event metrics collector gathers Kubernetes events from the cluster and sends them to the HakonGo API. The implementation includes:

1. An event collector that collects Kubernetes events
2. API client functionality to send event metrics to the HakonGo API
3. Mock API support for testing event metrics collection

## Test Files

The following test files have been implemented to verify the event metrics collection functionality:

### Event Collector Tests

Located at: `internal/collector/event_collector_test.go`

These tests verify that the event collector:
- Correctly collects events from the Kubernetes API
- Properly filters events based on namespace inclusion/exclusion rules
- Formats event data correctly with all required fields
- Properly categorizes events by severity (normal vs. warning)

### API Client Tests

Located at: `internal/api/client_test.go`

These tests verify that the API client:
- Correctly formats event metrics for sending to the HakonGo API
- Properly sends event metrics to the API endpoint
- Handles API errors correctly
- Filters out non-event resources

## Running the Tests

To run the event collector tests:

```bash
go test -v ./internal/collector -run TestEventCollector
```

To run the API client tests:

```bash
go test -v ./internal/api -run TestClient_SendEventMetrics
```

To run all tests:

```bash
go test ./...
```

## Mock API Testing

The mock API can be used to test the event metrics collection in a local environment. The mock API:

1. Receives event metrics at the `/v1/metrics/events` endpoint
2. Logs received event metrics to a file
3. Processes and displays event metrics in the dashboard UI

To deploy the mock API:

```bash
kubectl apply -f manifests/mock-api.yaml
```

Once deployed, you can access the mock API dashboard to view collected event metrics.

## Verifying Event Metrics Collection

To verify that event metrics are being collected and sent correctly:

1. Deploy the connector with event collection enabled
2. Generate some Kubernetes events (e.g., by deploying or deleting resources)
3. Check the connector logs for event collection messages
4. Access the mock API dashboard to view the collected event metrics

The dashboard will display:
- Total number of events collected
- Breakdown of events by severity (normal vs. warning)
- Raw event data for inspection

## Troubleshooting

If event metrics are not appearing in the dashboard:

1. Check the connector logs for any error messages related to event collection
2. Verify that the event collector is enabled in the connector configuration
3. Check the mock API logs for any errors processing event metrics
4. Ensure the connector has the necessary RBAC permissions to read events from the Kubernetes API
