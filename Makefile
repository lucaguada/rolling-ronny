# Rolling Ronny - build and publish helpers.  `make help` lists the targets.
GO      ?= go
REPO    ?= lucaguada/rolling-ronny
WEB     := web
WASM    := $(WEB)/rolling-ronny.wasm
PAGES_URL := https://$(word 1,$(subst /, ,$(REPO))).github.io/$(word 2,$(subst /, ,$(REPO)))/

.PHONY: help build run test vet web serve publish clean

help: ## show this help
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-8s %s\n", $$1, $$2}'

build: ## desktop binary ./rolling-ronny
	$(GO) build -o rolling-ronny .

run: ## run the desktop game
	$(GO) run .

vet: ## go vet
	$(GO) vet ./...

test: vet ## vet + unit tests
	$(GO) test ./...

web: $(WASM) $(WEB)/wasm_exec.js ## WebAssembly build into web/

$(WASM): $(wildcard *.go) go.mod
	GOOS=js GOARCH=wasm $(GO) build -ldflags="-s -w" -o $@ .
	@ls -lh $@

$(WEB)/wasm_exec.js:
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" $@

serve: web ## serve web/ on http://localhost:8080
	python3 -m http.server -d $(WEB) 8080

publish: test web ## force-push web/ to the gh-pages branch (GitHub Pages)
	@set -e; d=$$(mktemp -d); \
	cp $(WEB)/index.html $(WEB)/wasm_exec.js $(WASM) $$d/; touch $$d/.nojekyll; \
	cd $$d; git init -q -b gh-pages; git add -A; \
	git commit -q -m "Web build $$(date -u +%Y-%m-%dT%H:%MZ)" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"; \
	git -c http.postBuffer=524288000 -c credential.helper= -c credential.helper='!gh auth git-credential' \
	    push -q -f https://github.com/$(REPO).git gh-pages; \
	rm -rf $$d; echo "published: $(PAGES_URL)"

clean: ## remove build outputs
	rm -f rolling-ronny $(WASM) $(WEB)/wasm_exec.js
