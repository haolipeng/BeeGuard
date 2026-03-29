# Makefile for server server

# 变量定义
BINARY_NAME=server
GO=go
BUILD_DIR=build
MAIN_FILE=./cmd/main.go
MODULE=github.com/haolipeng/BeeGuard/server

# 部署目录
DEPLOY_DIR=/opt/cloudsec/server

# 版本信息（可通过命令行覆盖）
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 编译标志
LDFLAGS=-ldflags "-s -w \
	-X $(MODULE)/config.Version=$(VERSION) \
	-X $(MODULE)/config.BuildTime=$(BUILD_TIME) \
	-X $(MODULE)/config.GitCommit=$(GIT_COMMIT)"
GOFLAGS=-trimpath

# 默认目标
.PHONY: all
all: build

# 编译 server 主程序
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 清理编译产物
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# 部署到 /opt/cloudsec/server/
.PHONY: deploy
deploy: build
	@echo "Deploying $(BINARY_NAME) to $(DEPLOY_DIR)..."
	@sudo mkdir -p $(DEPLOY_DIR)/bin
	@sudo mkdir -p $(DEPLOY_DIR)/conf
	@sudo mkdir -p $(DEPLOY_DIR)/logs/server
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(DEPLOY_DIR)/bin/
	@sudo chmod 755 $(DEPLOY_DIR)/bin/$(BINARY_NAME)
	@sudo cp conf/server.yaml $(DEPLOY_DIR)/conf/server.yaml
	@echo "Config copied to $(DEPLOY_DIR)/conf/server.yaml"
	@sudo mkdir -p $(DEPLOY_DIR)/db_migrations
	@sudo cp -r db_migrations/*.sql $(DEPLOY_DIR)/db_migrations/
	@echo "Migrations copied to $(DEPLOY_DIR)/db_migrations/"
	@echo "Deploy complete: $(DEPLOY_DIR)/bin/$(BINARY_NAME)"

# 运行单元测试
.PHONY: test
test:
	@echo "Running unit tests..."
	$(GO) test ./... -v
	@echo "Tests complete"

# ========================
# 数据库迁移 (golang-migrate)
# ========================

# 数据库配置（可通过环境变量覆盖）
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASS ?= happy
DB_NAME ?= soc
DB_URL ?= postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS_DIR ?= db_migrations

# 执行所有未应用的迁移
.PHONY: migrate-up
migrate-up:
	@echo "Applying all migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "Migrations applied successfully"

# 回滚最后一次迁移
.PHONY: migrate-down
migrate-down:
	@echo "Rolling back last migration..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
	@echo "Rollback complete"

# 查看当前迁移版本
.PHONY: migrate-version
migrate-version:
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

# 重建数据库（回滚所有 + 重新应用）
.PHONY: migrate-reset
migrate-reset:
	@echo "Resetting database (drop all + migrate up)..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" drop -f
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "Database reset complete"

# 创建新迁移文件模板
.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq -digits 6 $$name
	@echo "Migration files created"

# 显示帮助信息
.PHONY: help
help:
	@echo "Server Makefile Commands:"
	@echo "  make build              - Build server binary"
	@echo "  make build VERSION=x.y.z - Build with specific version"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make deploy             - Deploy to $(DEPLOY_DIR)"
	@echo "  make test               - Run unit tests"
	@echo ""
	@echo "Database Migration Commands:"
	@echo "  make migrate-up         - Apply all pending migrations"
	@echo "  make migrate-down       - Rollback last migration"
	@echo "  make migrate-version    - Show current migration version"
	@echo "  make migrate-reset      - Drop all and re-migrate"
	@echo "  make migrate-create     - Create new migration files"
	@echo ""
	@echo "  make help               - Show this help message"
