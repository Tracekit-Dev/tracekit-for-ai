#!/bin/sh
set -eu
REPO_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac
BINARY="$REPO_DIR/bin/tracekit-agent-$OS-$ARCH"
if [ "$OS" = "mingw64_nt" ] || [ "$OS" = "msys_nt" ] || [ "$OS" = "cygwin_nt" ]; then
  BINARY="$BINARY.exe"
fi
if [ -x "$BINARY" ]; then
  exec "$BINARY" "$@"
fi
mkdir -p "$REPO_DIR/.cache/go-build"
cd "$REPO_DIR/agent/tracekit-agent"
exec env GOCACHE="$REPO_DIR/.cache/go-build" go run . "$@"
