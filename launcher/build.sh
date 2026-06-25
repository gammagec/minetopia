#!/usr/bin/env bash
# Build the launcher for all platforms.
set -euo pipefail

APP="minetopia-launcher"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS="-X main.Version=${VERSION} -s -w"

mkdir -p dist

build() {
  local os=$1 arch=$2 ext=${3:-}
  local out="dist/${APP}-${os}-${arch}${ext}"
  echo "Building $out ..."
  GOOS=$os GOARCH=$arch go build -ldflags="$LDFLAGS" -o "$out" .
}

cd "$(dirname "$0")"

build windows amd64  .exe
build windows arm64  .exe
build linux   amd64
build linux   arm64
build darwin  amd64
build darwin  arm64

echo ""
echo "Build complete:"
ls -lh dist/
