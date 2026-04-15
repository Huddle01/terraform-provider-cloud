.PHONY: fmt test build install-local

VERSION ?= 0.1.0
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)
LOCAL_PLUGIN_DIR := $(HOME)/.terraform.d/plugins/registry.terraform.io/huddle01/cloud/$(VERSION)/$(OS_ARCH)
LOCAL_PLUGIN_BIN := $(LOCAL_PLUGIN_DIR)/terraform-provider-cloud_v$(VERSION)

fmt:
	go fmt ./...

test:
	go test ./...

build:
	go build ./...

install-local:
	go build -o terraform-provider-cloud ./main.go
	mkdir -p "$(LOCAL_PLUGIN_DIR)"
	cp terraform-provider-cloud "$(LOCAL_PLUGIN_BIN)"
	@echo "Installed provider to $(LOCAL_PLUGIN_BIN)"
