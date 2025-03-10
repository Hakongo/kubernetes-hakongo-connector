FROM golang:1.24 as builder

# Build arguments
ARG VERSION=dev

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-X github.com/hakongo/kubernetes-connector/internal/version.Version=${VERSION}" -o manager cmd/controller/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

# Labels
LABEL org.opencontainers.image.source="https://github.com/hakongo/kubernetes-hakongo-connector"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.description="Kubernetes HakonGo Connector"
LABEL org.opencontainers.image.vendor="HakonGo"

ENTRYPOINT ["/manager"]
