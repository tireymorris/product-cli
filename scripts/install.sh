#!/usr/bin/env bash
set -euo pipefail

repo="${PRODUCT_CLI_REPO:-https://github.com/tireymorris/product-cli.git}"
ref="${1:-main}"

tmpdir="$(mktemp -d 2>/dev/null || mktemp -d -t product-cli-install)"
trap 'rm -rf "$tmpdir"' EXIT

git clone --depth 1 --branch "$ref" "$repo" "$tmpdir/product-cli"
(cd "$tmpdir/product-cli" && go install -ldflags "$(./scripts/version-ldflags.sh)" .)

gobin="$(go env GOBIN)"
if [[ -n "$gobin" ]]; then
  echo "installed ${gobin}/product-cli"
else
  echo "installed $(go env GOPATH)/bin/product-cli"
fi
