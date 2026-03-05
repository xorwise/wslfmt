#!/usr/bin/env sh
set -e

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
REPO="github.com/xorwise/wslfmt"

if ! command -v go >/dev/null 2>&1; then
  echo "error: go is not installed" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
GOBIN="$INSTALL_DIR" go install "${REPO}/cmd/wslfmt@latest"
echo "wslfmt installed to $INSTALL_DIR/wslfmt"
