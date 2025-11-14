#!/bin/bash

set -e

# Build directory in project root
BUILD_DIR="../build"

echo "📦 Creating build directory..."
mkdir -p "$BUILD_DIR"

# Try to copy wasm_exec.js from GOROOT, fallback to download
GOROOT=$(go env GOROOT)
WASM_EXEC="$GOROOT/misc/wasm/wasm_exec.js"

if [ -f "$WASM_EXEC" ]; then
  echo "✅ Found wasm_exec.js locally."
  cp "$WASM_EXEC" "$BUILD_DIR/"
else
  echo "⚠️ wasm_exec.js not found in $WASM_EXEC"
  echo "⬇️  Downloading wasm_exec.js from Golang GitHub..."
  curl -sSL -o "$BUILD_DIR/wasm_exec.js" https://raw.githubusercontent.com/golang/go/master/misc/wasm/wasm_exec.js
  echo "✅ Downloaded wasm_exec.js"
fi

echo "🔨 Building Go -> WASM..."
cd "$BUILD_DIR"
GOOS=js GOARCH=wasm go build -o sharedlib.wasm ../sharedlib/sharedlib_wasm.go
cd - > /dev/null

echo "✅ Build complete: build/sharedlib.wasm and wasm_exec.js saved to root build folder"
