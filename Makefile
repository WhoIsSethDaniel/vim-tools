GO := go
TARGETS := vim-check vim-add vim-remove vim-verify vim-list vim-enable vim-disable vim-build-sources vim-config vim-freeze vim-thaw vim-rename
PKG_TARGETS := $(TARGETS:%=./cmd/%)
BUILD_TARGETS := $(TARGETS:%=build/%)
BUILD_CMD := CGO_ENABLED=0 $(GO) build
INSTALL_CMD := CGO_ENABLED=0 $(GO) install

.PHONY: all
all:
	mkdir -p ./build
	$(BUILD_CMD) -o build $(patsubst %,./cmd/%,$(TARGETS))

$(BUILD_TARGETS): build/%:
	mkdir -p ./build
	$(BUILD_CMD) -o build ./cmd/$*

$(PKG_TARGETS): ./cmd/%: build/%

.PHONY: install
install:
	$(INSTALL_CMD) $(PKG_TARGETS)


.PHONY: $(TARGETS)
$(TARGETS): %: build/%

.PHONY: clean
clean:
	rm -f $(BUILD_TARGETS)
	rmdir ./build
