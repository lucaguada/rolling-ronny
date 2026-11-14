#!/bin/sh
# Builds the WebAssembly version into web/ (serve with: python3 -m http.server -d web 8080)
set -e
cd "$(dirname "$0")"
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/rolling-ronny.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/
ls -lh web/rolling-ronny.wasm
