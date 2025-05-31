BINARY_NAME := kbop

.PHONY: all
all: build

# 构建项目
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .

# 清理构建文件
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)

# 显示帮助信息
.PHONY: help
help:
	@echo "可用命令:"
	@echo "  make       或 make build  - 构建项目"
	@echo "  make clean               - 清理构建产物"
	@echo "  make help                - 帮助信息"