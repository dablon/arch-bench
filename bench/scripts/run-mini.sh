#!/bin/sh
# run-mini.sh: 20s × 50 VU on every cell.
# Emits: evidence/mini/<cell>.csv (per-request latency_ms)
#        evidence/mini/<cell>.summary

set -e
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
OUT="$ROOT/evidence/mini"
DURATION=${DURATION:-20}
mkdir -p "$OUT"
export JWT_SECRET="${JWT_SECRET:-bench-secret-9f3a}"

# Token — prefer $AB_BIN, fall back to in-source target, fall back to shared cargo-target.
TG_BIN=${AB_BIN:-}
[ -x "$TG_BIN" ] || TG_BIN="$ROOT/tools/token-gen/target/release/token-gen"
[ -x "$TG_BIN" ] || TG_BIN="$ROOT/.cargo-target/release/token-gen"
TOKEN=$(JWT_SECRET=$JWT_SECRET "$TG_BIN" --from-env --sub alice)
[ -n "$TOKEN" ] || { echo "token-gen failed at $TG_BIN" >&2; exit 1; }

# Resolve the binary path for a given cell. Layout varies:
#   go-flat-http:  ./go-flat-http
#   go-layered-http: ./bin/server
#   go-layered-uds: ./bin/server
#   go-layered-grpc: ./bin/server
#   go-hex-http: ./bin/server
#   go-hex-uds: ./bin/server
#   go-hex-grpc: ./bin/server
#   go-flat-uds / go-flat-grpc: ./go-flat-uds / ./go-flat-grpc
#   rust-*: ./target/release/<cell>
#   c-*: ./bin/<cell>
bin_for() {
    local cell="$1"
    case "$cell" in
        go-flat-*) echo "$ROOT/services/$cell/$cell" ;;
        go-*-http|go-*-uds|go-*-grpc)
            for cand in bin/server "$cell"; do
                [ -x "$ROOT/services/$cell/$cand" ] && echo "$ROOT/services/$cell/$cand" && return
            done
            echo "???"
            ;;
        rust-*)
            # Prefer the per-service target (in-source). Fall back to the shared
            # CARGO_TARGET_DIR used by the docker runner.
            for cand in "target/release/$cell" "../../.cargo-target/release/$cell"; do
                [ -x "$ROOT/services/$cell/$cand" ] && echo "$ROOT/services/$cell/$cand" && return
            done
            echo "???"
            ;;
        c-*) echo "$ROOT/services/$cell/bin/$cell" ;;
        *) echo "???" ;;
    esac
}

run_one() {
    local cell="$1"
    local kind="$2"
    local port="$3"
    local path="$4"
    local bin
    bin=$(bin_for "$cell")
    [ -x "$bin" ] || { echo "  missing $bin" >&2; return 1; }
    if [ "$kind" = "http" ]; then
        export LISTEN_ADDR="127.0.0.1:$port"
    elif [ "$kind" = "uds" ]; then
        export UDS_PATH="$path"
    else
        export GRPC_ADDR="127.0.0.1:$port"
    fi
    "$bin" >/dev/null 2>&1 &
    PID=$!
    sleep 0.8
    if [ "$kind" = "http" ]; then
        sh "$ROOT/bench/scripts/load-http.sh" "$port" "$TOKEN" "$DURATION" "$OUT/$cell.csv" > "$OUT/$cell.summary" 2>&1
    elif [ "$kind" = "uds" ]; then
        sh "$ROOT/bench/scripts/load-uds.sh" "$path" "$TOKEN" "$DURATION" > "$OUT/$cell.csv" 2> "$OUT/$cell.summary"
    else
        sh "$ROOT/bench/scripts/load-grpc.sh" "127.0.0.1:$port" "$TOKEN" "$DURATION" > "$OUT/$cell.csv" 2> "$OUT/$cell.summary"
    fi
    kill -9 $PID 2>/dev/null
    wait $PID 2>/dev/null
    sleep 0.2
}

# Cells: name:port (HTTP) or name:path (UDS) or name:port (gRPC)
HTTP_CELLS="go-flat-http:8080
go-layered-http:8081
go-hex-http:8082
rust-flat-http:8090
rust-layered-http:8091
rust-hex-http:8092
c-flat-http:9000
c-layered-http:9001
c-hex-http:9002"

UDS_CELLS="go-flat-uds:/tmp/go-flat-uds.sock
go-layered-uds:/tmp/go-layered-uds.sock
go-hex-uds:/tmp/go-hex-uds.sock
rust-flat-uds:/tmp/rust-flat-uds.sock
rust-layered-uds:/tmp/rust-layered-uds.sock
rust-hex-uds:/tmp/rust-hex-uds.sock
c-flat-uds:/tmp/c-flat-uds.sock
c-layered-uds:/tmp/c-layered-uds.sock
c-hex-uds:/tmp/c-hex-uds.sock"

GRPC_CELLS="go-flat-grpc:50051
go-layered-grpc:50052
go-hex-grpc:50053
rust-flat-grpc:51051
rust-layered-grpc:51052
rust-hex-grpc:51053"

for row in $HTTP_CELLS; do
    cell=${row%%:*}; port=${row##*:}
    echo ">>> $cell http ($port)"
    run_one "$cell" "http" "$port" "" || true
done

for row in $UDS_CELLS; do
    cell=${row%%:*}; path=${row#*:}
    echo ">>> $cell uds ($path)"
    run_one "$cell" "uds" "" "$path" || true
done

for row in $GRPC_CELLS; do
    cell=${row%%:*}; port=${row##*:}
    echo ">>> $cell grpc ($port)"
    run_one "$cell" "grpc" "$port" "" || true
done

echo "Done. See $OUT/"
