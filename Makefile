ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"


SERVER_NAME = "m3u8dl_for_web"
GOARCH ?= amd64
GOOS ?= linux
IMAGE_NAME = "ghcr.io/dorom741/m3u8dl_for_web"
GIT_HASH = $(shell git rev-parse --short  HEAD)
DOCKERFILE_PATH = Dockerfile


# DOCKER_USERNAME ?= your_docker_username
# DOCKER_PASSWORD ?= ${}
# DOCKER_REGISTRY ?= your_docker_registry


build_server:
	@echo "开始配置go环境..."
	go env -w GOSUMDB=off
	go mod download
#	@echo "开始检查代码..."
#	golangci-lint run --out-format=colored-line-number
	@echo "开始编译..."
	CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o ./bin/$(SERVER_NAME)


build_docker:
	docker build -t $(IMAGE_NAME):$(GIT_HASH) -t $(IMAGE_NAME):latest -f $(DOCKERFILE_PATH) . 


docker_push:
	echo $(DOCKER_PASSWORD) |  docker login -u $(DOCKER_USERNAME) --password-stdin  $(DOCKER_REGISTRY)
	docker push $(IMAGE_NAME):$(GIT_HASH)
	docker push $(IMAGE_NAME):latest