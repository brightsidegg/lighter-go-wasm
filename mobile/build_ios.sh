#!/bin/bash

set -e

# Build directory in project root
BUILD_DIR="../build"
FRAMEWORK_NAME="Lighter"

echo "🚀 Building iOS framework with gomobile..."
echo ""

# Check if gomobile is installed
if ! command -v gomobile &> /dev/null; then
    echo "❌ gomobile not found. Installing..."
    go install golang.org/x/mobile/cmd/gomobile@latest
    gomobile init
    echo "✅ gomobile installed"
fi

echo "📦 Creating build directory..."
mkdir -p "$BUILD_DIR"

echo ""
echo "🔨 Building xcframework for iOS + Simulator..."
echo "   This may take a few minutes on first build..."
echo ""

cd "$(dirname "$0")/.."

# Build xcframework that works on both iOS devices and simulator
# Use -mod=mod to avoid vendor issues with gomobile
GOFLAGS="-mod=mod" gomobile bind \
    -target=ios \
    -o "$BUILD_DIR/${FRAMEWORK_NAME}.xcframework" \
    -ldflags="-s -w" \
    -v \
    ./mobile

echo ""
echo "✅ Build complete!"
echo ""
echo "📱 Framework location: $BUILD_DIR/${FRAMEWORK_NAME}.xcframework"
echo ""
echo "🎯 To use in Xcode:"
echo "   1. Drag ${FRAMEWORK_NAME}.xcframework into your Xcode project"
echo "   2. In target settings, add to 'Frameworks, Libraries, and Embedded Content'"
echo "   3. Set to 'Embed & Sign'"
echo "   4. Import in Swift: import ${FRAMEWORK_NAME}"
echo ""
echo "📖 Example Swift usage:"
echo '   ```swift'
echo '   import Lighter'
echo ''
echo '   // Generate API key'
echo '   let result = MobileGenerateAPIKey("")'
echo '   if result?.error == "" {'
echo '       print("Private Key: \(result!.privateKey!)")'
echo '       print("Public Key: \(result!.publicKey!)")'
echo '   }'
echo ''
echo '   // Create client'
echo '   let err = MobileCreateClient('
echo '       "https://api.lighter.xyz",'
echo '       result!.privateKey,'
echo '       42,    // chainId'
echo '       0,     // apiKeyIndex'
echo '       123    // accountIndex'
echo '   )'
echo '   if err == "" {'
echo '       print("Client created successfully!")'
echo '   }'
echo '   ```'
echo ""

