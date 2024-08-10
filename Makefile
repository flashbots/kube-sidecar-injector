VERSION := $(shell git describe --tags --always --dirty="-dev" --match "v*.*.*" || echo "development" )
VERSION := $(VERSION:v%=%)

.PHONY: build
build:
	@CGO_ENABLED=0 go build \
			-ldflags "-X main.version=${VERSION}" \
			-o ./bin/kube-sidecar-injector \
		github.com/flashbots/kube-sidecar-injector/cmd

.PHONY: docker
docker:
	docker build -t kube-sidecar-injector:${VERSION} .

.PHONY: snapshot
snapshot:
	@goreleaser release --snapshot --clean

.PHONY: image
image:
	@docker build \
			--build-arg VERSION=${VERSION} \
			--tag kube-sidecar-injector:${VERSION} \
		.

.PHONY: deploy
deploy: image
	@kubectl \
			--context orbstack \
		apply \
			--filename test/cluster-role.yaml \
			--filename test/configmap.yaml \
			--filename test/dummy.yaml \
			--filename test/deployment-fargate.yaml \
			--filename test/deployment-node-exporter.yaml
