ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"


SERVER_NAME = "m3u8dl"
GOARCH ?= amd64
GOOS ?= linux
IMAGE_NAME = "ghcr.io/dorom741/m3u8dl_for_web"
GIT_HASH = $(shell git rev-parse --short  HEAD)
DOCKERFILE_PATH = Dockerfile
DOCKERFILE_CUDA_PATH = Dockerfile.cuda


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
	CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o ./bin/$(SERVER_NAME) ./cmd

build_arm64_linux:
	@echo "开始编译..."
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o ./bin/$(SERVER_NAME)_linux_arm64 ./cmd


build_docker:
	docker build \
		--build-arg HTTP_PROXY="$${HTTP_PROXY:-$${http_proxy}}" \
		--build-arg HTTPS_PROXY="$${HTTPS_PROXY:-$${https_proxy}}" \
		--build-arg NO_PROXY="$${NO_PROXY:-$${no_proxy}}" \
		--build-arg http_proxy="$${http_proxy:-$${HTTP_PROXY}}" \
		--build-arg https_proxy="$${https_proxy:-$${HTTPS_PROXY}}" \
		--build-arg no_proxy="$${no_proxy:-$${NO_PROXY}}" \
		-t $(IMAGE_NAME):$(GIT_HASH) -t $(IMAGE_NAME):latest -f $(DOCKERFILE_PATH) .

build_docker_cuda:
	docker build \
		--build-arg HTTP_PROXY="$${HTTP_PROXY:-$${http_proxy}}" \
		--build-arg HTTPS_PROXY="$${HTTPS_PROXY:-$${https_proxy}}" \
		--build-arg NO_PROXY="$${NO_PROXY:-$${no_proxy}}" \
		--build-arg http_proxy="$${http_proxy:-$${HTTP_PROXY}}" \
		--build-arg https_proxy="$${https_proxy:-$${HTTPS_PROXY}}" \
		--build-arg no_proxy="$${no_proxy:-$${NO_PROXY}}" \
		-t $(IMAGE_NAME):$(GIT_HASH)-cuda -t $(IMAGE_NAME):latest-cuda -f $(DOCKERFILE_CUDA_PATH) .


docker_push:
	echo $(DOCKER_PASSWORD) |  docker login -u $(DOCKER_USERNAME) --password-stdin  $(DOCKER_REGISTRY)
	docker push $(IMAGE_NAME):$(GIT_HASH)
	docker push $(IMAGE_NAME):latest

docker_push_cuda:
	echo $(DOCKER_PASSWORD) |  docker login -u $(DOCKER_USERNAME) --password-stdin  $(DOCKER_REGISTRY)
	docker push $(IMAGE_NAME):$(GIT_HASH)-cuda
	docker push $(IMAGE_NAME):latest-cuda
