#!/bin/sh
# entrypoint-bench.sh — build every service, run run-mini.sh, then aggregate.
# Usage: entrypoint-bench.sh [DURATION] [OUT_DIR]
set -e
DURATION=${1:-20}
OUT=${2:-/src/evidence/mini}
echo "=== Building all 24 services for the bench ==="
mkdir -p "$OUT"

# Build each Go service. Flat: ./<name>. Layered/hex: ./bin/server.
for s in go-flat-http go-flat-uds go-flat-grpc; do
    echo "build $s"
    (cd /src/services/$s && go build -o /src/services/$s/$s .) || { echo "  FAIL $s"; exit 1; }
done
for s in go-layered-http go-layered-uds go-layered-grpc go-hex-http go-hex-uds go-hex-grpc; do
    echo "build $s"
    (cd /src/services/$s/cmd/server && go build -o /src/services/$s/bin/server .) || { echo "  FAIL $s"; exit 1; }
done

# Build gRPC load generator.
echo "build grpc-load"
(cd /src/bench/scripts/grpc-load && go build -o /tmp/grpc-load .) || { echo "  FAIL grpc-load"; exit 1; }
ls -la /tmp/grpc-load

# Build each Rust service (release).
for s in rust-flat-http rust-flat-uds rust-flat-grpc rust-layered-http rust-layered-uds rust-layered-grpc rust-hex-http rust-hex-uds rust-hex-grpc; do
    echo "build $s"
    (cd /src/services/$s && cargo build --release --bin $s) || { echo "  FAIL $s"; exit 1; }
done

# Build each C service. Output at bin/<name>.
for s in c-flat-http c-flat-uds c-layered-http c-layered-uds c-hex-http c-hex-uds; do
    echo "build $s"
    (cd /src/services/$s && make 2>&1) || { echo "  FAIL $s"; exit 1; }
done

# Build token-gen.
(cd /src/tools/token-gen && cargo build --release 2>&1 | tail -3)
export AB_BIN=/src/.cargo-target/release/token-gen
ls -la $AB_BIN || { echo "token-gen build failed"; exit 2; }

echo ""
echo "=== Verifying binaries ==="
for s in /src/services/c-*/bin/* /src/services/rust-*/target/release/rust-* /src/services/go-flat-* /src/services/go-*/bin/server; do
    [ -x $s ] && echo "  OK $s" || echo "  MISSING $s"
done

echo ""
echo "=== Running mini-bench DURATION=${DURATION}s ==="
export DURATION
cd /src
sh bench/scripts/run-mini.sh 2>&1 | tail -40
echo ""
echo "=== Aggregating ==="
python3 bench/scripts/aggregate.py "$OUT" > /src/evidence/comparison.md 2>&1 || true
cat /src/evidence/comparison.md | head -40
