BINARY_NAME=csi-shared-lvm
IMAGE_NAME=csi-shared-lvm
IMAGE_TAG=latest

.PHONY: all build test clean image

all: build

build:
	go build -o bin/$(BINARY_NAME) main.go

test:
	go test ./...

image:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

clean:
	go clean
	rm -rf bin/
