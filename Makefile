BINPATH ?= build

build:
	go build -tags 'production' -o $(BINPATH)/dp-frontend-dataset-controller

debug:
	go build -tags 'debug' -o $(BINPATH)/dp-frontend-dataset-controller
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-dataset-controller

test:
	go test -tags 'production' ./...

.PHONY: build debug
