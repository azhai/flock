# ── Project config ────────────────────────────────────────

# main binary name (built from ./)
APP       = gouache
# frontend directory
ADMIN_DIR = web
# auto-discover cmd/ subdirs
COMMANDS  := $(notdir $(patsubst %/.,%,$(wildcard cmd/*/.)))

# ── Build flags ───────────────────────────────────────────

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
RELEASE    = -ldflags "-s -w -X main.version=$(VERSION)"
GOBUILD    = go build $(RELEASE)

# ── Targets ───────────────────────────────────────────────

.PHONY: one all front build run dev clean test test-watch cover server lint tidy mk

## one: local single-platform build (CGO disabled, current OS/ARCH)
one: front
	@echo "Build $(APP) (local) ..."
	mkdir -p bin
	CGO_ENABLED=0 $(GOBUILD) -o bin/$(APP) ./
	@for cmd in $(COMMANDS); do \
		CGO_ENABLED=0 $(GOBUILD) -o bin/$$cmd ./cmd/$$cmd; \
	done

## all: full pipeline — clean, then local build, then cross-compile
all: clean one build

## build: cross-compile APP + all COMMANDS for 5 platforms
build: front
	@echo "Cross-compiling ..."
	mkdir -p bin
	@for target in $(APP) $(COMMANDS); do \
		if [ "$$target" = "$(APP)" ]; then src=.; else src=./cmd/$$target; fi; \
		CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -o bin/$$target-$(VERSION).darwin-arm64 $$src && \
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/$$target-$(VERSION).darwin-amd64 $$src && \
		CGO_ENABLED=0 GOOS=linux  GOARCH=arm64 $(GOBUILD) -o bin/$$target-$(VERSION).linux-arm64  $$src && \
		CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 $(GOBUILD) -o bin/$$target-$(VERSION).linux-amd64  $$src && \
		if [ "$$target" != "$(APP)" ]; then \
			CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$$target-$(VERSION).exe  $$src; \
		fi; \
	done
	@echo "✅ Build success."

## front: build frontend assets (tsc type-check + vite build)
front:
	@echo "Build frontend ..."
	cd $(ADMIN_DIR) && npm run build

## run: build frontend then run main with go run
run: front
	go run ./

## server: 编译后端并启动（仅编译 + 运行二进制，不依赖前端构建）
server: one
	./bin/$(APP)

## dev: start hot-reload backend (air) + frontend dev server (vite)
dev:
	@which air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	@if [ ! -d "$(ADMIN_DIR)/node_modules" ]; then cd $(ADMIN_DIR) && npm install; fi
	(cd $(ADMIN_DIR) && npm run dev &)
	air

## mk: alias for dev (hot reload)
mk: dev

## clean: remove build artifacts and frontend dist
clean:
	rm -rf bin/* tmp/* $(ADMIN_DIR)/dist
	@echo "✅ Clean complete."

## test: 运行 Go 测试（verbose + race + 覆盖率聚合），可 TAGS=... 透传
test:
	go test -race -coverpkg=./... ./... -coverprofile=/tmp/gouache.cover $(TAGS)
	@go tool cover -func=/tmp/gouache.cover | tail -1

## test-watch: 热重载测试 —— 编辑任意 .go 文件后自动重跑并输出覆盖率
test-watch:
	@which air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	air -c .air.test.toml

## cover: 生成覆盖率报告并在浏览器打开
cover:
	@go test -coverpkg=./... ./... -coverprofile=/tmp/gouache.cover > /dev/null && go tool cover -html=/tmp/gouache.cover

## lint: run golangci-lint on entire project
lint:
	golangci-lint run ./...

## tidy: clean up go.mod dependencies
tidy:
	go mod tidy
