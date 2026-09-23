APP_NAME   := agentpet
APP_BUNDLE := dist/$(APP_NAME).app
ICON_SRC   := asserts/icon.png
PLIST_SRC  := asserts/Info.plist
ICONSET    := dist/AppIcon.iconset

##@ Build
build: ## Build the agentpet binary.
	mkdir -p dist
	go build -o dist/$(APP_NAME) .

app: ## Package agentpet as a macOS .app bundle.
	@mkdir -p $(APP_BUNDLE)/Contents/MacOS \
	           $(APP_BUNDLE)/Contents/Resources \
	           $(ICONSET)
	go build -o $(APP_BUNDLE)/Contents/MacOS/$(APP_NAME) .
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

.PHONY: build app test clean help
