#!/bin/sh
set -e
DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$DIR"
make -s clean
make -s
PORT=19302
export JWT_SECRET="test"
export LISTEN_ADDR="127.0.0.1:$PORT"
./bin/c-hex-http >/tmp/chext.log 2>&1 &
PID=$!
sleep 0.5
trap 'kill -9 $PID 2>/dev/null || true; wait $PID 2>/dev/null || true' EXIT
TOKEN=$(JWT_SECRET=$JWT_SECRET ${AB_BIN:-$HOME/arch-bench/tools/token-gen/target/release/token-gen} --from-env --sub alice)
BAD="garbage.garbage.garbage"
http() { curl -s -m 5 -w "\n%{http_code}" "$@"; }
R=$(http http://127.0.0.1:$PORT/health); [ "$(echo "$R"|tail -1)" = "200" ] || { echo FAIL health; exit 1; }
R=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"$TOKEN\"}" http://127.0.0.1:$PORT/verify)
B=$(echo "$R"|head -n -1); C=$(echo "$R"|tail -1)
[ "$C" = "200" ] || { echo FAIL ok; exit 1; }
echo "$B" | grep -q '"OK"' || { echo FAIL body; exit 1; }
echo "$B" | grep -q '"alice"' || { echo FAIL body; exit 1; }
R=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"$BAD\"}" http://127.0.0.1:$PORT/verify)
[ "$(echo "$R"|tail -1)" = "401" ] || { echo FAIL bad; exit 1; }
R=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"\"}" http://127.0.0.1:$PORT/verify)
[ "$(echo "$R"|tail -1)" = "400" ] || { echo FAIL empty; exit 1; }
R=$(http -X POST -H "Content-Type: application/json" -d "not json" http://127.0.0.1:$PORT/verify)
[ "$(echo "$R"|tail -1)" = "400" ] || { echo FAIL badbody; exit 1; }
R=$(http -X GET http://127.0.0.1:$PORT/verify)
[ "$(echo "$R"|tail -1)" = "405" ] || { echo FAIL method; exit 1; }
EXP=$(JWT_SECRET=$JWT_SECRET ${AB_BIN:-$HOME/arch-bench/tools/token-gen/target/release/token-gen} --from-env --sub bob --ttl 1)
sleep 2.5
R=$(http -X POST -H "Content-Type: application/json" -d "{\"token\":\"$EXP\"}" http://127.0.0.1:$PORT/verify)
[ "$(echo "$R"|tail -1)" = "401" ] || { echo FAIL expired; exit 1; }
echo "PASS: c-hex-http all tests green"
