#!/bin/sh
# Tests for c-flat-http.
set -e
DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$DIR"

# Build
make -s clean
make -s

# Start service
PORT=19300
export JWT_SECRET="test"
export LISTEN_ADDR="127.0.0.1:$PORT"
./bin/c-flat-http >/tmp/cflat-test.log 2>&1 &
PID=$!
sleep 2.5

cleanup() {
  kill -9 $PID 2>/dev/null || true
  wait $PID 2>/dev/null || true
}
trap cleanup EXIT

TOKEN=$(JWT_SECRET=$JWT_SECRET /workspace/projects/arch-bench/tools/token-gen/target/release/token-gen --from-env --sub alice)
BAD_TOKEN="garbage.garbage.garbage"

# Helper: curl with timeout
http() {
  curl -s -m 5 -w "\n%{http_code}" "$@"
}

# 1. Health
RESP=$(http http://127.0.0.1:$PORT/health)
CODE=$(echo "$RESP" | tail -1)
[ "$CODE" = "200" ] || { echo "FAIL: health got $CODE"; cat /tmp/cflat-test.log; exit 1; }

# 2. Verify OK
RESP=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"$TOKEN\"}" http://127.0.0.1:$PORT/verify)
CODE=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | head -n -1)
[ "$CODE" = "200" ] || { echo "FAIL: verify ok got $CODE body=$BODY"; exit 1; }
echo "$BODY" | grep -q '"OK"' || { echo "FAIL: body=$BODY"; exit 1; }
echo "$BODY" | grep -q '"alice"' || { echo "FAIL: body=$BODY"; exit 1; }

# 3. Verify bad token
RESP=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"$BAD_TOKEN\"}" http://127.0.0.1:$PORT/verify)
CODE=$(echo "$RESP" | tail -1)
[ "$CODE" = "401" ] || { echo "FAIL: bad token got $CODE"; exit 1; }

# 4. Verify empty token
RESP=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"\"}" http://127.0.0.1:$PORT/verify)
CODE=$(echo "$RESP" | tail -1)
[ "$CODE" = "400" ] || { echo "FAIL: empty got $CODE"; exit 1; }

# 5. Verify bad body
RESP=$(http -X POST -H "Content-Type: application/json" -d "not json" http://127.0.0.1:$PORT/verify)
CODE=$(echo "$RESP" | tail -1)
[ "$CODE" = "400" ] || { echo "FAIL: bad body got $CODE"; exit 1; }

# 6. Verify bad method
RESP=$(http -X GET http://127.0.0.1:$PORT/verify)
CODE=$(echo "$RESP" | tail -1)
[ "$CODE" = "405" ] || { echo "FAIL: bad method got $CODE"; exit 1; }

# 7. Verify expired - generate a short-lived token, wait for it to expire
EXP_TOKEN=$(JWT_SECRET=$JWT_SECRET /workspace/projects/arch-bench/tools/token-gen/target/release/token-gen --from-env --sub bob --ttl 1)
sleep 2.5
RESP=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"$EXP_TOKEN\"}" http://127.0.0.1:$PORT/verify)
CODE=$(echo "$RESP" | tail -1)
[ "$CODE" = "401" ] || { echo "FAIL: expired got $CODE"; exit 1; }

echo "PASS: c-flat-http all tests green"
