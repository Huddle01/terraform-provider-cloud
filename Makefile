.PHONY: fmt test test-acceptance build install-local dev-override

VERSION ?= 0.1.0
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)
LOCAL_PLUGIN_DIR := $(HOME)/.terraform.d/plugins/registry.terraform.io/huddle01/cloud/$(VERSION)/$(OS_ARCH)
LOCAL_PLUGIN_BIN := $(LOCAL_PLUGIN_DIR)/terraform-provider-cloud_v$(VERSION)

# Directory used by dev-override (version-independent binary path)
DEV_PLUGIN_DIR := $(HOME)/.terraform.d/plugins-dev
DEV_PLUGIN_BIN := $(DEV_PLUGIN_DIR)/terraform-provider-cloud
TERRAFORMRC     := $(HOME)/.terraformrc

fmt:
	go fmt ./...

test:
	go test ./...

# test-acceptance runs end-to-end tests against a real Huddle API.
# Required env vars: HUDDLE_API_KEY, HUDDLE_REGION
# Instance/attachment tests also need: HUDDLE_FLAVOR_ID, HUDDLE_IMAGE_ID
# Optional: HUDDLE_BASE_URL, HUDDLE_SSH_PUBLIC_KEY
test-acceptance:
	TF_ACC=1 go test ./... -run TestAcc -v -timeout 30m

build:
	go build ./...

install-local:
	go build -o terraform-provider-cloud ./main.go
	mkdir -p "$(LOCAL_PLUGIN_DIR)"
	cp terraform-provider-cloud "$(LOCAL_PLUGIN_BIN)"
	@echo "Installed provider to $(LOCAL_PLUGIN_BIN)"

# dev-override: build the provider and configure ~/.terraformrc to use the local
# binary via dev_overrides. This bypasses registry version resolution entirely, so
# you never need to bump VERSION to match the registry during local development.
#
# Usage:
#   make dev-override          # one-time setup
#   cd examples/vm-with-volume && terraform plan  # skip `terraform init`
#
# To stop using the override, remove the dev_overrides block from ~/.terraformrc.
dev-override:
	go build -o "$(DEV_PLUGIN_BIN)" ./main.go
	@mkdir -p "$(DEV_PLUGIN_DIR)"
	@if grep -q 'dev_overrides' "$(TERRAFORMRC)" 2>/dev/null; then \
		echo "dev_overrides block already present in $(TERRAFORMRC) — update it manually if the path changed."; \
	else \
		printf '\nprovider_installation {\n  dev_overrides {\n    "huddle01/cloud" = "$(DEV_PLUGIN_DIR)"\n  }\n  direct {}\n}\n' >> "$(TERRAFORMRC)"; \
		echo "Added dev_overrides to $(TERRAFORMRC)"; \
	fi
	@echo "Built provider to $(DEV_PLUGIN_BIN)"
	@echo "Run 'terraform plan' directly in your example dir (skip terraform init)."
