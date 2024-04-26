VERSION := $(shell git describe --tags --always --dirty="-dev" --match "v*.*.*" || echo "development" )
VERSION := $(VERSION:v%=%)

.PHONY: build
build:
	@CGO_ENABLED=0 go build \
			-ldflags "-X main.version=${VERSION}" \
			-o ./bin/kube-sidecar-injector \
		github.com/flashbots/kube-sidecar-injector/cmd

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
deploy:
	@kubectl \
			--context orbstack \
		apply \
			--filename deploy/cluster-role.yaml \
			--filename deploy/dummy.yaml \
			--filename deploy/deployment-fargate.yaml \
			--filename deploy/deployment-node-exporter.yaml
