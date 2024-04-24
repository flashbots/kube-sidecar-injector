# stage: build ---------------------------------------------------------

FROM golang:1.22-alpine as build

RUN apk add --no-cache gcc musl-dev linux-headers

WORKDIR /go/src/github.com/flashbots/kube-sidecar-injector

COPY go.* ./
RUN go mod download

COPY . .

ARG VERSION

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build \
            -o bin/kube-sidecar-injector \
            -ldflags "-s -w -X main.version=${VERSION}" \
        github.com/flashbots/kube-sidecar-injector/cmd

# stage: run -----------------------------------------------------------

# TODO: change for distroless

FROM alpine

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /go/src/github.com/flashbots/kube-sidecar-injector/bin/kube-sidecar-injector ./kube-sidecar-injector

ENTRYPOINT ["/app/kube-sidecar-injector"]
