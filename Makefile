GO := go
TARGETS := vim-check vim-add vim-remove vim-verify vim-list vim-enable vim-disable
PKG_TARGETS := $(TARGETS:%=./cmd/%)
BUILD_TARGETS := $(TARGETS:%=build/%)

.PHONY: all
all: 
	mkdir -p ./build
	$(GO) build -o build $(patsubst %,./cmd/%,$(TARGETS))

$(BUILD_TARGETS): build/%: 
	mkdir -p ./build
	$(GO) build -o build ./cmd/$*

$(PKG_TARGETS): ./cmd/%: build/%

.PHONY: install
install: all
	$(GO) install $(PKG_TARGETS)


.PHONY: $(TARGETS)
$(TARGETS): %: build/%

.PHONY: clean
clean:
	rm -f $(BUILD_TARGETS)
	rmdir ./build
