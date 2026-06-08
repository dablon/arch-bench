#!/bin/bash
# Automated Benchmark Runner for arch-bench
set -e
BASE_DIR="/home/nalcaraz/arch-bench"
EVIDENCE_DIR="$BASE_DIR/evidence/benchmark-results"
mkdir -p "$EVIDENCE_DIR"
TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
RESULTS_FILE="$EVIDENCE_DIR/results_$TIMESTAMP.json"
log() { echo -e "\033[0;36m[BM]\033[0m $1"; }
cleanup() {
    log "Cleaning up..."
    pkill -f "rust-flat-http" 2>/dev/null || true
    pkill -f "rust-hex-http" 2>/dev/null || true
    pkill -f "rust-layered-http" 2>/dev/null || true
    pkill -f "go-flat-http" 2>/dev/null || true
    pkill -f "go-hex-http" 2>/dev/null || true
    pkill -f "go-layered-http" 2>/dev/null || true
    sleep 2
}
cleanup
log "Starting benchmark automation..."

start_http() {
    local name=$1; local binary=$2; local port=$3
    log "Starting $name:$port..."
    rm -f /tmp/${name}.sock 2>/dev/null
    if [[ "$binary" == *"/bin/server"* ]]; then
        (cd $(dirname $binary) && env JWT_SECRET="***" LISTEN_ADDR="127.0.0.1:$port" ./bin/server > /tmp/${name}.log 2>&1 &)
    else
        (env JWT_SECRET="***" LISTEN_ADDR="127.0.0.1:$port" $binary > /tmp/${name}.log 2>&1 &)
    fi
    sleep 3
    ss -tlnp | grep -q ":$port " && echo "OK: $name" || echo "FAIL: $name"
}
