# syntax=docker/dockerfile:1
FROM golang:1.25 AS builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY cmd/ cmd/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -a -o csi-shared-lvm main.go

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    lvm2 \
    e2fsprogs \
    xfsprogs \
    util-linux \
    udev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /
COPY --from=builder /workspace/csi-shared-lvm .

CMD ["/csi-shared-lvm"]
