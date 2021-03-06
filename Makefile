
# Image URL to use all building/pushing image targets
IMG ?= thewebroot/volrec:v0.0.1
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go \
	-set-owner \
    -set-ns

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Deploy controller using the "Prod" example overlay
deploy-prod: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/overlays/prod | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# Restart all the pods
restart: 
	kubectl rollout restart deploy -n volrec-system volrec-controller

# Do all the things
build: docker-build docker-push deploy restart

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Setup Test Namespace and artifacts
test-setup:
	kubectl create ns test1
	kubectl label ns test1 k8s.twr.dev/owner="user1"
	kubectl apply -f ./testing/kubernetes/crdb-sts.yaml -n test1

set-reclaim-delete:
	kubectl -n test1 label pvc --all storage.k8s.twr.dev/reclaim-policy=Delete --overwrite

set-reclaim-retain:
	kubectl -n test1 label pvc --all storage.k8s.twr.dev/reclaim-policy=Retain --overwrite

set-reclaim-recycle:
	kubectl -n test1 label pvc --all storage.k8s.twr.dev/reclaim-policy=Recycle --overwrite

