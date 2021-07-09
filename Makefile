BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -m all | nancy sleuth

.PHONY: build
build: generate-prod
	go build -tags 'production' -o $(BINPATH)/dp-frontend-dataset-controller -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

.PHONY: debug
debug: generate-debug
	go build -tags 'debug' -o $(BINPATH)/dp-frontend-dataset-controller -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-dataset-controller

.PHONY: test
test: generate-prod
	go test -race -cover -tags 'production' ./...

.PHONY: convey
convey:
	goconvey ./...

.PHONY: all build debug audit

.PHONY: get-renderer-version
get-renderer-version:
	$(eval APP_RENDERER_VERSION=$(shell grep "github.com/ONSdigital/dp-renderer" go.mod | cut -d ' ' -f2 ))

.PHONY: fetch-renderer-lib
fetch-renderer-lib: get-renderer-version
	$(eval CORE_ASSETS_PATH = $(shell go get github.com/ONSdigital/dp-renderer@$(APP_RENDERER_VERSION) && go list -f '{{.Dir}}' -m github.com/ONSdigital/dp-renderer))

.PHONY: generate-debug
generate-debug: fetch-renderer-lib
	cd assets; go run github.com/kevinburke/go-bindata/go-bindata -prefix $(CORE_ASSETS_PATH)/assets -debug -o data.go -pkg assets locales/... templates/... $(CORE_ASSETS_PATH)/assets/locales/... $(CORE_ASSETS_PATH)/assets/templates/...
	{ echo "// +build debug\n"; cat assets/data.go; } > assets/debug.go.new
	mv assets/debug.go.new assets/data.go

.PHONY: generate-prod
generate-prod: fetch-renderer-lib 
	cd assets; go run github.com/kevinburke/go-bindata/go-bindata -prefix $(CORE_ASSETS_PATH)/assets -o data.go -pkg assets locales/... templates/... $(CORE_ASSETS_PATH)/assets/locales/... $(CORE_ASSETS_PATH)/assets/templates/...
	{ echo "// +build production\n"; cat assets/data.go; } > assets/data.go.new
	mv assets/data.go.new assets/data.go
