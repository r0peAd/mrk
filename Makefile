APP := conncheck
PACKAGE := ./cmd/conncheck
BUILD_DIR := build

TARGETS := darwin_amd64 darwin_arm64 \
           linux_amd64 linux_arm64 linux_386 linux_arm \
           windows_amd64 windows_arm64

LDFLAGS := -s -w

.PHONY: all build build-all checksums test fmt vet clean help

all: test build

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(APP) $(PACKAGE)

build-all: $(addprefix build-,$(TARGETS))

checksums:
	cd $(BUILD_DIR) && shasum -a 256 */$(APP) */$(APP).exe > SHA256SUMS

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR) $(APP)

help:
	@echo "Usage:"
	@echo "  make build             Build for the current OS and architecture"
	@echo "  make build-all         Build all supported OS and architecture combinations"
	@echo "  make checksums         Generate SHA256SUMS for all cross builds"
	@echo "  make test              Run tests"
	@echo "  make fmt               Format source code"
	@echo "  make vet               Run go vet"
	@echo "  make clean             Remove build outputs"

define CROSS_BUILD
.PHONY: build-$(1)
build-$(1):
	@mkdir -p $(BUILD_DIR)/$(1)
	@GOOS=$(word 1,$(subst _, ,$(1))) \
		GOARCH=$(word 2,$(subst _, ,$(1))) \
		CGO_ENABLED=0 \
		go build -trimpath -ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(1)/$(if $(filter windows,$(word 1,$(subst _, ,$(1)))),$(APP).exe,$(APP)) \
		$(PACKAGE)
	@echo "Built $(BUILD_DIR)/$(1)"
endef

$(foreach target,$(TARGETS),$(eval $(call CROSS_BUILD,$(target))))
