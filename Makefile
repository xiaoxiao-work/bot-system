# 应用和镜像配置
NAME := bot-system
VERSION ?= latest
REGISTRY := 192.168.31.37:5000
REPO := $(REGISTRY)/$(NAME)
TAG := $(REPO):$(VERSION)

# Go编译参数
GOOS ?= linux
GOARCH ?= amd64
LDFLAGS := -s -w

.PHONY: all
all: push

# 编译应用
.PHONY: build
build:
	@echo "编译 Bot Manager..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(LDFLAGS)" -o "$(NAME)" .
	@echo "编译完成: $(NAME)"

# 构建镜像
.PHONY: image
image: build
	@echo "构建镜像 $(TAG)..."
	docker build -t $(TAG) -f Dockerfile .
	@if [ "$(VERSION)" != "latest" ]; then \
		docker tag $(TAG) $(REPO):latest; \
		echo "已添加 latest 标签"; \
	fi
	@echo "镜像构建完成: $(TAG)"

# 推送镜像（构建 + 推送）
.PHONY: push
push: image
	@echo "推送镜像到仓库..."
	docker push $(TAG)
	@if [ "$(VERSION)" != "latest" ]; then \
		docker push $(REPO):latest; \
		echo "已推送 latest 标签"; \
	fi
	@echo "推送完成: $(TAG)"
	@echo "清理本地二进制文件..."
	@rm -f $(NAME)

# 本地运行（开发用）
.PHONY: run
run:
	@echo "编译并运行 Bot Manager..."
	go run .

# 清理
.PHONY: clean
clean:
	@echo "清理编译产物..."
	@rm -f $(NAME) *.exe
	@echo "清理完成"

# 显示帮助
.PHONY: help
help:
	@echo "Bot Manager Makefile 命令:"
	@echo ""
	@echo "  make push [VERSION=x.y.z]  - 构建并推送镜像到仓库（默认）"
	@echo "  make build                 - 仅编译应用"
	@echo "  make image [VERSION=x.y.z] - 仅构建镜像"
	@echo "  make run                   - 本地运行（开发用）"
	@echo "  make clean                 - 清理编译产物"
	@echo ""
	@echo "示例:"
	@echo "  推送 latest 版本:  make push"
	@echo "  推送指定版本:      make push VERSION=1.0.0"
	@echo ""
	@echo "镜像仓库: $(REGISTRY)"
