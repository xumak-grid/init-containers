.PHONY: build

IMAGE_NAME = /grid/init-nexus
IMAGE_TAG = 1.0.0
FULL_IMAGE_NAME = $(IMAGE_NAME):$(IMAGE_TAG)

default: build

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/init-nexus main.go config.go
	docker build -t $(FULL_IMAGE_NAME) --no-cache .
push:
	docker push $(FULL_IMAGE_NAME)
