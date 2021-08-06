SHELL = bash

PROJECT_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
THIS_OS := $(shell uname | cut -d- -f1)
THIS_ARCH := $(shell uname -m)

GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_DIRTY := $(if $(shell git status --porcelain),+CHANGES)

GO_LDFLAGS := "-X github.com/angrycub/parse-nomad-config/version.GitCommit=$(GIT_COMMIT)$(GIT_DIRTY)"

GO_TAGS ?= nonvidia

ifeq ($(CI),true)
GO_TAGS := codegen_generated $(GO_TAGS)
endif

GO_TEST_CMD = $(if $(shell command -v gotestsum 2>/dev/null),gotestsum --,go test)

ifeq ($(origin GOTEST_PKGS_EXCLUDE), undefined)
GOTEST_PKGS ?= "./..."
else
GOTEST_PKGS=$(shell go list ./... | sed 's/github.com\/hashicorp\/parse-nomad-config/./' | egrep -v "^($(GOTEST_PKGS_EXCLUDE))(/.*)?$$")
endif

default: help

ifeq ($(CI),true)
	$(info Running in a CI environment, verbose mode is disabled)
else
	VERBOSE="true"
endif

ifeq (Darwin,$(THIS_OS))
ALL_TARGETS = linux_amd64 \
	darwin_amd64 \
	linux_arm64 \
	darwin_arm64 
endif

ifeq (Linux,$(THIS_OS))
ALL_TARGETS = linux_amd64 \
	darwin_amd64 \
	linux_arm64 \
	darwin_arm64 \
	freebsd_amd64 \
	windows_amd64 windows_amd86 \
	linux_arm linux_s390x linux_386
endif


OTHER_TARGETS = windows_amd64 windows_amd86 linux_arm linux_s390x linux_386 freebsd_amd64


SUPPORTED_OSES = Darwin Linux FreeBSD Windows MSYS_NT

# include per-user customization after all variables are defined
-include GNUMakefile.local

pkg/%/parse-nomad-config: GO_OUT ?= $@
#pkg/%/parse-nomad-config: CC ?= $(shell go env CC)
pkg/%/parse-nomad-config: ## Build Nomad for GOOS_GOARCH, e.g. pkg/linux_amd64/parse-nomad-config
ifeq (,$(findstring $(THIS_OS),$(SUPPORTED_OSES)))
	$(warning WARNING: Building Nomad is only supported on $(SUPPORTED_OSES); not $(THIS_OS))
endif
	@echo "==> Building $@ with tags $(GO_TAGS)..."
	@CGO_ENABLED=0 \
		GOOS=$(firstword $(subst _, ,$*)) \
		GOARCH=$(lastword $(subst _, ,$*)) \
		go build -trimpath -ldflags $(GO_LDFLAGS) -tags "$(GO_TAGS)" -o $(GO_OUT)

pkg/windows_%/parse-nomad-config: GO_OUT = $@.exe

# Define package targets for each of the build targets we actually have on this system
define makePackageTarget

pkg/$(1).zip: pkg/$(1)/parse-nomad-config
	@echo "==> Packaging for $(1)..."
	@zip -j pkg/$(1).zip pkg/$(1)/*

endef

# Reify the package targets
$(foreach t,$(ALL_TARGETS),$(eval $(call makePackageTarget,$(t))))

.PHONY: bootstrap
bootstrap: deps lint-deps git-hooks # Install all dependencies

.PHONY: deps
deps:  ## Install build and development dependencies
	@echo "==> Updating build dependencies..."
## maybe install gox
# go install github.com/hashicorp/go-bindata/go-bindata@bf7910af899725e4938903fb32048c7c0b15f12e


.PHONY: lint-deps
lint-deps: ## Install linter dependencies
## Keep versions in sync with tools/go.mod (see https://github.com/golang/go/issues/30515)
	@echo "==> Updating linter dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.39.0
	go install github.com/client9/misspell/cmd/misspell@v0.3.4
	go install github.com/hashicorp/go-hclog/hclogvet@v0.1.3


.PHONY: check
check: ## Lint the source code
	@echo "==> Linting source code..."
	@golangci-lint run -j 1

	@echo "==> Linting hclog statements..."
	@hclogvet .

	@echo "==> Spell checking website..."
	@misspell -error -source=text website/pages/

	@echo "==> Checking Go mod.."
	@GO111MODULE=on $(MAKE) tidy
	@if (git status --porcelain | grep -Eq "go\.(mod|sum)"); then \
		echo go.mod or go.sum needs updating; \
		git --no-pager diff go.mod; \
		git --no-pager diff go.sum; \
		exit 1; fi

	@echo "==> Check raft util msg type mapping are in-sync..."
	@go generate ./helper/raftutil/
	@if (git status -s ./helper/raftutil| grep -q .go); then echo "raftutil helper message type mapping is out of sync. Run go generate ./... and push."; exit 1; fi


.PHONY: tidy
tidy:
	@echo "--> Tidy parse-nomad-config module"
	@go mod tidy

.PHONY: dev
dev: GOOS=$(shell go env GOOS)
dev: GOARCH=$(shell go env GOARCH)
dev: GOPATH=$(shell go env GOPATH)
dev: DEV_TARGET=pkg/$(GOOS)_$(GOARCH)/parse-nomad-config
dev: ## Build for the current development platform
	@echo "==> Removing old development build..."
	@rm -f $(PROJECT_ROOT)/$(DEV_TARGET)
	@rm -f $(PROJECT_ROOT)/bin/parse-nomad-config
	@rm -f $(GOPATH)/bin/parse-nomad-config
	@$(MAKE) --no-print-directory \
		$(DEV_TARGET) \
		GO_TAGS="$(GO_TAGS) $(NOMAD_UI_TAG)"
	@mkdir -p $(PROJECT_ROOT)/bin
	@mkdir -p $(GOPATH)/bin
	@cp $(PROJECT_ROOT)/$(DEV_TARGET) $(PROJECT_ROOT)/bin/
	@cp $(PROJECT_ROOT)/$(DEV_TARGET) $(GOPATH)/bin

.PHONY: release
release: GO_TAGS=release nonvidia
release: clean $(foreach t,$(ALL_TARGETS),pkg/$(t).zip) ## Build all release packages which can be built on this platform.
	@echo "==> Results:"
	@tree --dirsfirst $(PROJECT_ROOT)/pkg

.PHONY: test
test: ## Run the parse-nomad-config test suite
	@if [ ! $(SKIP_TESTS) ]; then \
		make test-nomad; \
		fi

.PHONY: test-nomad
test-nomad: dev ## Run Nomad test suites
	@echo "==> Running Nomad test suites:"
	$(if $(ENABLE_RACE),GORACE="strip_path_prefix=$(GOPATH)/src") $(GO_TEST_CMD) \
		$(if $(ENABLE_RACE),-race) $(if $(VERBOSE),-v) \
		-cover \
		-timeout=15m \
		-tags "$(GO_TAGS)" \
		$(GOTEST_PKGS) $(if $(VERBOSE), >test.log ; echo $$? > exit-code)
	@if [ $(VERBOSE) ] ; then \
		bash -C "$(PROJECT_ROOT)/scripts/test_check.sh" ; \
	fi

.PHONY: clean
clean: GOPATH=$(shell go env GOPATH)
clean: ## Remove build artifacts
	@echo "==> Cleaning build artifacts..."
	@rm -rf "$(PROJECT_ROOT)/bin/*"
	@echo "==> Cleaning $(PROJECT_ROOT)/pkg/* "
	@rm -rf "$(PROJECT_ROOT)/pkg/*"
	@rm -f "$(GOPATH)/bin/parse-nomad-config"

HELP_FORMAT="    \033[36m%-25s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo ""
	@echo "This host will build the following targets if 'make release' is invoked:"
	@echo $(ALL_TARGETS) | sed 's/^/    /'
