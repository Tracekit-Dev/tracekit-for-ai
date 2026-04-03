#!/bin/sh
set -eu
REPO_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
exec "$REPO_DIR/scripts/run-tracekit-agent.sh" mcp "$@"
