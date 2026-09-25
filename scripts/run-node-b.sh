#!/usr/bin/env bash
# Build and run node B. Same code as node A, different identity and endpoint:
# the two are told apart by configuration, not by their source.
set -euo pipefail

cd "$(dirname "$0")/.."

go build -o bin/node ./cmd/node

exec ./bin/node -id node-b -http 127.0.0.1:8082
