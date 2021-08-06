GOARCH    ?= $(shell go env GOARCH)
GOOS      ?= $(shell go env GOOS)

PKG             := github.com/fieldju/awesom-o
SRC_DIRS        := cmd
BUILD_DIR       := ${PWD}/dist/$(GOOS)_$(GOARCH)

.PHONY: build-dirs
build-dirs:
	@mkdir -p $(BUILD_DIR)

.PHONY: build
build: build-dirs
	@echo "Building Awesom-0 ..."
	@go build -o ${BUILD_DIR}/awesom-o main.go
	@echo "Beep. Boop. I will be your best friend"