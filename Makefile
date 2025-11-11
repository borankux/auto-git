.PHONY: build install clean reinstall

BINARY_NAME=auto-git
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_INSTALL=$(GO_CMD) install
GO_CLEAN=$(GO_CMD) clean
GO_TEST=$(GO_CMD) test
GO_GET=$(GO_CMD) get

build:
	@echo "Building $(BINARY_NAME)..."
	$(GO_BUILD) -o bin/$(BINARY_NAME) .

install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO_INSTALL) .
	@echo "Creating 'ag' symlink..."
	@GOBIN=$$(go env GOBIN); \
	if [ -z "$$GOBIN" ]; then \
		GOPATH=$$(go env GOPATH); \
		if [ -z "$$GOPATH" ]; then \
			GOPATH=$$HOME/go; \
		fi; \
		GOBIN="$$GOPATH/bin"; \
	fi; \
	if [ -f "$$GOBIN/auto-git" ]; then \
		ln -sf "$$GOBIN/auto-git" "$$GOBIN/ag" 2>/dev/null || true; \
		echo "Symlink created: $$GOBIN/ag -> $$GOBIN/auto-git"; \
		echo "Make sure $$GOBIN is in your PATH"; \
	else \
		echo "Warning: auto-git binary not found in $$GOBIN"; \
	fi

clean:
	@echo "Cleaning..."
	$(GO_CLEAN)
	rm -f bin/$(BINARY_NAME)

reinstall: clean build install
	@echo "Cleaned, built, and installed $(BINARY_NAME)."

test:
	@echo "Running tests..."
	$(GO_TEST) ./...

deps:
	@echo "Downloading dependencies..."
	$(GO_GET) -d ./...
