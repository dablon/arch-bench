#!/bin/sh
# entrypoint-test.sh — run tests for all (or a subset of) services.
# Usage: entrypoint-test.sh [--all | --go | --rust | --c | <service-name>]

set -e
PASS=0
FAIL=0
FAILED_LIST=""

# Build token-gen once so c test.sh can find it.
echo "=== Building token-gen ==="
(cd /src/tools/token-gen && cargo build --release 2>&1 | tail -3)
# token-gen is built with the shared CARGO_TARGET_DIR=/src/.cargo-target
export AB_BIN=/src/.cargo-target/release/token-gen
ls -la $AB_BIN || { echo "token-gen build failed (tried $AB_BIN)"; exit 2; }
echo ""

# Each test invocation captures stdout+stderr+exit. Prints PASS/FAIL summary.
run_go() {
    name=$1
    cwd=$2
    extra=${3:-}
    cd "$cwd"
    out=$(go test ./... $extra 2>&1) && { echo "PASS $name"; PASS=$((PASS+1)); } \
        || { echo "FAIL $name"; echo "$out" | tail -5; FAIL=$((FAIL+1)); FAILED_LIST="$FAILED_LIST $name"; }
}

run_rust() {
    name=$1
    cwd=$2
    cd "$cwd"
    out=$(cargo test --release --lib 2>&1) && { echo "PASS $name"; PASS=$((PASS+1)); } \
        || { echo "FAIL $name"; echo "$out" | tail -5; FAIL=$((FAIL+1)); FAILED_LIST="$FAILED_LIST $name"; }
}

run_c() {
    name=$1
    cwd=$2
    cd "$cwd"
    out=$(sh tests/test.sh 2>&1) && { echo "PASS $name"; PASS=$((PASS+1)); } \
        || { echo "FAIL $name"; echo "$out" | tail -5; FAIL=$((FAIL+1)); FAILED_LIST="$FAILED_LIST $name"; }
}

test_service() {
    name=$1
    case "$name" in
        go-*)
            run_go "$name" "/src/services/$name"
            ;;
        rust-*)
            run_rust "$name" "/src/services/$name"
            ;;
        c-*)
            run_c "$name" "/src/services/$name"
            ;;
        *)
            echo "UNKNOWN $name"
            ;;
    esac
}

target=${1:---all}

case "$target" in
    --all)
        for svc in /src/services/*/; do
            test_service "$(basename $svc)"
        done
        ;;
    --go) for s in /src/services/go-*/; do test_service "$(basename $s)"; done ;;
    --rust) for s in /src/services/rust-*/; do test_service "$(basename $s)"; done ;;
    --c) for s in /src/services/c-*/; do test_service "$(basename $s)"; done ;;
    --grpc)
        for s in /src/services/*-grpc/; do test_service "$(basename $s)"; done
        ;;
    --http)
        for s in /src/services/*-http/; do test_service "$(basename $s)"; done
        ;;
    --uds)
        for s in /src/services/*-uds/; do test_service "$(basename $s)"; done
        ;;
    --parallel)
        # Run Go, Rust, and C groups in parallel, then aggregate.
        (for s in /src/services/go-*/; do test_service "$(basename $s)"; done) &
        P1=$!
        (for s in /src/services/rust-*/; do test_service "$(basename $s)"; done) &
        P2=$!
        (for s in /src/services/c-*/; do test_service "$(basename $s)"; done) &
        P3=$!
        wait $P1; wait $P2; wait $P3
        ;;
    *)
        test_service "$target"
        ;;
esac

echo ""
echo "============================================"
echo "RESULTS: $PASS passed, $FAIL failed"
if [ -n "$FAILED_LIST" ]; then
    echo "FAILED:$FAILED_LIST"
    exit 1
fi
echo "ALL TESTS GREEN"
