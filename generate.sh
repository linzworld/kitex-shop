#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
KITEX_BIN=${KITEX_BIN:-}

if [[ -z "$KITEX_BIN" ]]; then
    if command -v kitex >/dev/null 2>&1; then
        KITEX_BIN=$(command -v kitex)
    else
        KITEX_BIN=$(go env GOPATH)/bin/kitex
    fi
fi

if [[ ! -x "$KITEX_BIN" ]]; then
    echo "kitex executable not found; install it with: go install github.com/cloudwego/kitex/tool/cmd/kitex@v0.16.3" >&2
    exit 1
fi

cd "$ROOT_DIR"
rm -rf "$ROOT_DIR/kitex_gen"

"$KITEX_BIN" -module example_shop -I idl idl/base.proto
"$KITEX_BIN" -module example_shop -I idl idl/stock.proto
"$KITEX_BIN" -module example_shop -I idl idl/item.proto
