#!/bin/sh
# gRPC load generator (Go binary).
# Usage: load-grpc.sh <addr> <token> <duration_sec>

set -e
ADDR="$1"; TOKEN="$2"; DURATION="$3"
LOAD=/tmp/grpc-load
[ -x "$LOAD" ] || { echo "missing $LOAD; build it first" >&2; exit 1; }
"$LOAD" -addr "$ADDR" -token "$TOKEN" -duration "$DURATION"
