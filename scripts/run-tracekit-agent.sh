#!/bin/sh
set -eu
REPO_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
RELEASE_BASE_URL="https://github.com/Tracekit-Dev/tracekit-for-ai/releases/latest/download"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac
BINARY="$REPO_DIR/bin/tracekit-agent-$OS-$ARCH"
SHA_FILE="$BINARY.sha256"
if [ "$OS" = "mingw64_nt" ] || [ "$OS" = "msys_nt" ] || [ "$OS" = "cygwin_nt" ]; then
  OS="windows"
  BINARY="$BINARY.exe"
  SHA_FILE="$SHA_FILE.exe.sha256"
fi

download_binary() {
  mkdir -p "$REPO_DIR/bin"
  BINARY_NAME=$(basename "$BINARY")
  SHA_NAME=$(basename "$SHA_FILE")

  echo "Downloading $BINARY_NAME from the latest TraceKit release..." >&2
  rm -f "$BINARY" "$SHA_FILE"
  curl -fsSL "$RELEASE_BASE_URL/$BINARY_NAME" -o "$BINARY"
  curl -fsSL "$RELEASE_BASE_URL/$SHA_NAME" -o "$SHA_FILE"

  EXPECTED_HASH=$(awk 'NR==1 {print $1}' "$SHA_FILE")
  if [ -z "$EXPECTED_HASH" ]; then
    echo "Downloaded checksum file did not contain a SHA256 hash." >&2
    rm -f "$BINARY" "$SHA_FILE"
    exit 1
  fi

  if command -v shasum >/dev/null 2>&1; then
    ACTUAL_HASH=$(shasum -a 256 "$BINARY" | awk '{print $1}')
  elif command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_HASH=$(sha256sum "$BINARY" | awk '{print $1}')
  else
    echo "No SHA256 verification tool found (expected shasum or sha256sum)." >&2
    rm -f "$BINARY" "$SHA_FILE"
    exit 1
  fi

  if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then
    echo "SHA256 verification failed for $BINARY_NAME." >&2
    rm -f "$BINARY" "$SHA_FILE"
    exit 1
  fi

  chmod +x "$BINARY"
}

if [ -x "$BINARY" ]; then
  exec "$BINARY" "$@"
fi

if command -v curl >/dev/null 2>&1; then
  download_binary
  exec "$BINARY" "$@"
fi

mkdir -p "$REPO_DIR/.cache/go-build"
cd "$REPO_DIR/agent/tracekit-agent"
exec env GOCACHE="$REPO_DIR/.cache/go-build" go run . "$@"
