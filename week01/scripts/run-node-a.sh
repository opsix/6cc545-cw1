#!/bin/sh
# Build and run node A.
#
# The binary is built first and then executed, rather than using "go run". With
# "go run" the process you interrupt is the go tool, not the node, so the
# shutdown path being demonstrated is not the node's. Building first means
# Ctrl-C reaches the node itself.
set -eu

cd "$(dirname "$0")/.."

go build -o bin/node ./cmd/node

exec ./bin/node -id node-a -http 127.0.0.1:8081
