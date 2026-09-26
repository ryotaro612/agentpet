APP_NAME   := agentpet
APP_BUNDLE := dist/$(APP_NAME).app
BINARY     := dist/$(APP_NAME)
CLIENT     := dist/petowner
ICON_SRC   := asserts/icon.png
PLIST_SRC  := asserts/Info.plist
ICONSET    := dist/AppIcon.iconset
ASSETS     := $(ICON_SRC) $(PLIST_SRC)

GO_SRCS  := main.go $(shell find internal -name '*.go')
CMD_SRCS := $(wildcard cmd/*.go)

##@ Build
all: petowner app ## Build both the petowner client and the app bundle.

build: $(BINARY) ## Build the agentpet binary.

$(BINARY): $(GO_SRCS) | dist
	go build -o $@ .

petowner: $(CLIENT) ## Build the petowner client binary.

$(CLIENT): $(CMD_SRCS) | dist ## Build the petowner client binary.
	go build -o $@ ./cmd/

app: $(BINARY) $(ASSETS) ## Package agentpet as a macOS .app bundle.
	@mkdir -p $(APP_BUNDLE)/Contents/MacOS \
	           $(APP_BUNDLE)/Contents/Resources \
	           $(ICONSET)
	@cp $(BINARY) $(APP_BUNDLE)/Contents/MacOS/$(APP_NAME)
	@sips -z 16  16  $(ICON_SRC) --out $(ICONSET)/icon_16x16.png      > /dev/null
	@sips -z 32  32  $(ICON_SRC) --out $(ICONSET)/icon_16x16@2x.png   > /dev/null
	@sips -z 32  32  $(ICON_SRC) --out $(ICONSET)/icon_32x32.png      > /dev/null
	@sips -z 64  64  $(ICON_SRC) --out $(ICONSET)/icon_32x32@2x.png   > /dev/null
	@sips -z 128 128 $(ICON_SRC) --out $(ICONSET)/icon_128x128.png    > /dev/null
	@sips -z 256 256 $(ICON_SRC) --out $(ICONSET)/icon_128x128@2x.png > /dev/null
	@sips -z 256 256 $(ICON_SRC) --out $(ICONSET)/icon_256x256.png    > /dev/null
	@sips -z 512 512 $(ICON_SRC) --out $(ICONSET)/icon_256x256@2x.png > /dev/null
	@sips -z 512 512 $(ICON_SRC) --out $(ICONSET)/icon_512x512.png    > /dev/null
	@cp              $(ICON_SRC)       $(ICONSET)/icon_512x512@2x.png
	@iconutil -c icns $(ICONSET) -o $(APP_BUNDLE)/Contents/Resources/AppIcon.icns
	@rm -rf $(ICONSET)
	@cp $(PLIST_SRC) $(APP_BUNDLE)/Contents/Info.plist
	@echo "Built $(APP_BUNDLE)"

dist:
	@mkdir -p dist

##@ Test
test: ## Run tests.
	go test ./...

##@ Clean
clean: ## Remove generated files.
	rm -rf build dist

##@ Help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

.PHONY: all build petowner app test clean help dist
