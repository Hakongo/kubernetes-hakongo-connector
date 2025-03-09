# Image URL to use all building/pushing image targets
VERSION ?= $(shell git describe --tags --always --dirty)
REGISTRY ?= ghcr.io
ORG ?= hakongo
IMAGE_NAME ?= kubernetes-hakongo-connector
IMG ?= $(REGISTRY)/$(ORG)/$(IMAGE_NAME):$(VERSION)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.29.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'.
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

.PHONY: generate
generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	controller-gen object paths="./api/..."

##@ Build

.PHONY: manifests
manifests: ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	controller-gen rbac:roleName=manager-role crd webhook paths="./api/..." output:crd:artifacts:config=config/crd/bases

.PHONY: build
build: fmt vet generate manifests ## Build manager binary.
	go build -o bin/manager cmd/controller/main.go

.PHONY: run
run: fmt vet generate manifests ## Run a controller from your host.
	go run ./cmd/controller/main.go

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} --build-arg VERSION=$(VERSION) .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

.PHONY: docker-buildx
docker-buildx: ## Build and push multi-platform docker image with buildx
	docker buildx build --platform linux/amd64,linux/arm64 --push -t ${IMG} --build-arg VERSION=$(VERSION) .

.PHONY: release
release: docker-build docker-push ## Build and push a release

##@ Deployment

.PHONY: deploy
deploy: ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	@sed -e 's|ghcr.io/hakongo/kubernetes-hakongo-connector:latest|$(IMG)|g' manifests/manager.yaml | kubectl apply -f -
	kubectl apply -f manifests/mock-api.yaml

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f manifests/
