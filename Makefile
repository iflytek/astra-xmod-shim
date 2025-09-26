# ===================================
# astra-xmod-shim - Makefile
# 生产级构建、容器化、发布与 Helm 部署
# 使用: make build | make docker | make push | make docker-multiarch | make helm-package
# ===================================

# --- 默认目标 ---
.DEFAULT_GOAL := help

# --- 项目变量 ---
BINARY = astra-xmod-shim
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT_ID ?= $(shell git rev-parse --short HEAD)
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

# --- 输出路径 ---
OUTPUT = ./bin/$(BINARY)

# --- 镜像配置 ---
REGISTRY ?= ghcr.io
IMAGE_REPO ?= $(REGISTRY)/iflytek/$(BINARY)
IMAGE_TAG ?= $(VERSION)

# --- Helm Chart 目录 ---
HELM_CHART_DIR := deploy/helm/$(BINARY)

# --- Go 编译参数 ---
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT_ID) -X main.branch=$(BRANCH)"
GO_ENV := GO111MODULE=on

.PHONY: build docker push run clean help

# --- 核心目标 ---

## 编译二进制
build:
	@echo "Building $(BINARY) v$(VERSION) [$(BRANCH)/$(COMMIT_ID)]"
	@$(GO_ENV) go build $(LDFLAGS) -o $(OUTPUT) cmd/server/main.go
	@echo "Build complete: $(OUTPUT)"

## 构建 Docker 镜像（本地开发使用）
docker: build
	@echo "Building Docker image: $(IMAGE_REPO):$(IMAGE_TAG)"
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT_ID=$(COMMIT_ID) \
		--build-arg BRANCH=$(BRANCH) \
		-t $(IMAGE_REPO):$(IMAGE_TAG) \
		-t $(IMAGE_REPO):latest \
		-f deploy/docker/Dockerfile .
	@echo "Docker image built"

## 推送镜像到 registry（需先登录）
push: docker
	@echo "Pushing $(IMAGE_REPO):$(IMAGE_TAG) and :latest to $(REGISTRY)"
	@docker push $(IMAGE_REPO):$(IMAGE_TAG)
	@docker push $(IMAGE_REPO):latest
	@echo "Images pushed"

## 构建并推送多架构镜像（CI 使用，支持 amd64 + arm64）
.PHONY: docker-multiarch
docker-multiarch:
	@echo "Building multi-arch image for linux/amd64,linux/arm64"
	@docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--push \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT_ID=$(COMMIT_ID) \
		--build-arg BRANCH=$(BRANCH) \
		-t $(IMAGE_REPO):$(IMAGE_TAG) \
		-t $(IMAGE_REPO):latest \
		-f deploy/docker/Dockerfile .
	@echo "Multi-arch images pushed to $(IMAGE_REPO)"

## 本地运行（用于调试）
run:
	@echo "Running $(BINARY)..."
	@$(GO_ENV) go run cmd/server/main.go -c conf/base/conf.yaml

## 清理构建产物
clean:
	@echo "Cleaning up..."
	@rm -rf ./bin/ dist/
	@echo "Clean done"

# --- 开发工具命令 ---

## 格式化代码
.PHONY: fmt
fmt:
	@$(GO_ENV) go fmt ./...
	@echo "Code formatted"

## 运行 Go vet 检查
.PHONY: vet
vet:
	@$(GO_ENV) go vet ./...
	@echo "Go vet completed"

## 运行代码检查
.PHONY: lint
lint:
	@golangci-lint run
	@echo "Linting completed"

## 运行测试
.PHONY: test
test:
	@$(GO_ENV) go test ./... -v
	@echo "Tests completed"

# --- Helm 相关命令 ---

## 打包 Helm Chart
.PHONY: helm-package
helm-package:
	@echo "Packaging Helm chart..."
	@mkdir -p dist
	@helm package $(HELM_CHART_DIR) --version $(VERSION) --destination dist/
	@echo "Helm chart packaged: dist/$(BINARY)-$(VERSION).tgz"

## 安装 Helm Chart
.PHONY: helm-install
helm-install:
	@echo "Installing Helm chart..."
	@helm install $(BINARY) $(HELM_CHART_DIR) --namespace default --create-namespace
	@echo "Helm chart installed"

## 卸载 Helm Chart
.PHONY: helm-uninstall
helm-uninstall:
	@echo "Uninstalling Helm chart..."
	@helm uninstall $(BINARY) --namespace default
	@echo "Helm chart uninstalled"

# --- 便捷命令 ---

## 构建所有内容（二进制 + 镜像 + Helm 包）
.PHONY: all
all: build docker helm-package

## 开发模式（格式化 → 检查 → 测试 → 运行）
.PHONY: dev
dev: fmt vet lint test run

# --- 帮助信息 ---
help:
	@echo ""
	@echo "astra-xmod-shim Makefile"
	@echo ""
	@echo "构建与发布:"
	@echo "  make build               # 编译二进制"
	@echo "  make docker              # 构建本地镜像"
	@echo "  make push                # 构建并推送镜像到 registry"
	@echo "  make docker-multiarch    # CI: 构建并推送多架构镜像"
	@echo "  make all                 # 构建所有内容"
	@echo ""
	@echo "开发与测试:"
	@echo "  make dev                 # 开发流程：fmt → vet → lint → test → run"
	@echo "  make fmt                 # 格式化代码"
	@echo "  make vet                 # Go vet 检查"
	@echo "  make lint                # 静态检查"
	@echo "  make test                # 运行测试"
	@echo "  make run                 # 本地运行"
	@echo "  make clean               # 清理 bin/ 和 dist/"
	@echo ""
	@echo "Helm 部署:"
	@echo "  make helm-package        # 打包 Helm chart 到 dist/"
	@echo "  make helm-install        # 安装 Helm chart"
	@echo "  make helm-uninstall      # 卸载 Helm chart"
	@echo ""
	@echo "提示：可设置 VERSION=1.0.0 或 REGISTRY=ghcr.io 自定义行为"
	@echo ""