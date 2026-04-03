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

needs_update() {
  if [ ! -x "$BINARY" ]; then
    return 0
  fi
  command -v curl >/dev/null 2>&1 || return 1
  CHECK_FILE="$REPO_DIR/bin/.last-update-check"
  if [ -f "$CHECK_FILE" ]; then
    LAST_CHECK=$(cat "$CHECK_FILE" 2>/dev/null || echo 0)
    NOW=$(date +%s)
    ELAPSED=$((NOW - LAST_CHECK))
    if [ "$ELAPSED" -lt 3600 ]; then
      return 1
    fi
  fi
  LOCAL_VERSION=$("$BINARY" version 2>/dev/null || echo "unknown")
  LATEST_VERSION=$(curl -fsSL "https://api.github.com/repos/Tracekit-Dev/tracekit-for-ai/releases/latest" 2>/dev/null | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)
  mkdir -p "$REPO_DIR/bin"
  date +%s > "$CHECK_FILE"
  if [ -z "$LATEST_VERSION" ] || [ "$LATEST_VERSION" = "$LOCAL_VERSION" ]; then
    return 1
  fi
  echo "Update available: $LOCAL_VERSION -> $LATEST_VERSION" >&2
  return 0
}

if [ -x "$BINARY" ]; then
  if needs_update; then
    download_binary
  fi
  exec "$BINARY" "$@"
fi

if command -v curl >/dev/null 2>&1; then
  download_binary
  exec "$BINARY" "$@"
fi

mkdir -p "$REPO_DIR/.cache/go-build"
cd "$REPO_DIR/agent/tracekit-agent"
exec env GOCACHE="$REPO_DIR/.cache/go-build" go run . "$@"
