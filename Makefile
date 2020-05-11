BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

all: audit test build

audit:
	nancy go.sum

build:
	go build -tags 'production' -o $(BINPATH)/dp-frontend-dataset-controller -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

debug:
	go build -tags 'debug' -o $(BINPATH)/dp-frontend-dataset-controller -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-dataset-controller

test:
	go test -race -cover ./...

convey:
	goconvey ./...

.PHONY: all build debug audit
