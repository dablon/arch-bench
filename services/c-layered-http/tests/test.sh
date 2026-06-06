#!/bin/sh
# Tests for c-layered-http.
set -e
DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$DIR"

make -s clean
make -s

PORT=19301
export JWT_SECRET="test"
export LISTEN_ADDR="127.0.0.1:$PORT"
./bin/c-layered-http >/tmp/clayered.log 2>&1 &
PID=$!
sleep 0.5
trap 'kill -9 $PID 2>/dev/null || true; wait $PID 2>/dev/null || true' EXIT

TOKEN=$(JWT_SECRET=$JWT_SECRET /workspace/projects/arch-bench/tools/token-gen/target/release/token-gen --from-env --sub alice)
BAD_TOKEN="garbage.garbage.garbage"

http() { curl -s -m 5 -w "\n%{http_code}" "$@"; }

RESP=$(http http://127.0.0.1:$PORT/health)
[ "$(echo "$RESP" | tail -1)" = "200" ] || { echo FAIL health; cat /tmp/clayered.log; exit 1; }

RESP=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"$TOKEN\"}" http://127.0.0.1:$PORT/verify)
BODY=$(echo "$RESP" | head -n -1); CODE=$(echo "$RESP" | tail -1)
[ "$CODE" = "200" ] || { echo FAIL ok; exit 1; }
echo "$BODY" | grep -q '"OK"' || { echo FAIL body; exit 1; }
echo "$BODY" | grep -q '"alice"' || { echo FAIL body; exit 1; }

RESP=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"$BAD_TOKEN\"}" http://127.0.0.1:$PORT/verify)
[ "$(echo "$RESP" | tail -1)" = "401" ] || { echo FAIL bad; exit 1; }

RESP=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"\"}" http://127.0.0.1:$PORT/verify)
[ "$(echo "$RESP" | tail -1)" = "400" ] || { echo FAIL empty; exit 1; }

RESP=$(http -X POST -H "Content-Type: application/json" -d "not json" http://127.0.0.1:$PORT/verify)
[ "$(echo "$RESP" | tail -1)" = "400" ] || { echo FAIL badbody; exit 1; }

RESP=$(http -X GET http://127.0.0.1:$PORT/verify)
[ "$(echo "$RESP" | tail -1)" = "405" ] || { echo FAIL method; exit 1; }

EXP_TOKEN=$(JWT_SECRET=$JWT_SECRET /workspace/projects/arch-bench/tools/token-gen/target/release/token-gen --from-env --sub bob --ttl 1)
sleep 2.5
RESP=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"$EXP_TOKEN\"}" http://127.0.0.1:$PORT/verify)
[ "$(echo "$RESP" | tail -1)" = "401" ] || { echo FAIL expired; exit 1; }

echo "PASS: c-layered-http all tests green"
