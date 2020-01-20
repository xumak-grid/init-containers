.PHONY: build push

IMAGE_NAME = /grid/init-gogs
IMAGE_TAG = 1.0.0
FULL_IMAGE_NAME = $(IMAGE_NAME):$(IMAGE_TAG)

default: build

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o ./bin/init-gogs *.go
	docker build -t $(FULL_IMAGE_NAME) --no-cache .
push:	
	docker push $(FULL_IMAGE_NAME)
