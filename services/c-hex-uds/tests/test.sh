#!/bin/sh
set -e
DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$DIR"
make -s clean
make -s
export JWT_SECRET="test"
export UDS_PATH="/tmp/c-hex-uds-test.sock"
rm -f "$UDS_PATH"
./bin/c-hex-uds >/tmp/chex-uds.log 2>&1 &
PID=$!
sleep 0.5
trap 'kill -9 $PID 2>/dev/null || true; wait $PID 2>/dev/null || true; rm -f "$UDS_PATH"' EXIT
TOKEN=$(JWT_SECRET=$JWT_SECRET ${AB_BIN:-$HOME/arch-bench/tools/token-gen/target/release/token-gen} --from-env --sub alice)
BAD="garbage.garbage.garbage"
PY=/tmp/uds_client.py
cat > "$PY" << 'EOF'
import socket, sys
s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
s.connect(sys.argv[1])
s.sendall((sys.argv[2] + "\n").encode())
buf = b""
while not buf.endswith(b"\n"):
    chunk = s.recv(1024)
    if not chunk: break
    buf += chunk
print(buf.decode().rstrip())
EOF
R=$(python3 "$PY" "$UDS_PATH" "VERIFY $TOKEN"); [ "$R" = "OK alice" ] || { echo "FAIL ok: $R"; exit 1; }
R=$(python3 "$PY" "$UDS_PATH" "VERIFY $BAD"); [ "$R" = "ERR INVALID_TOKEN" ] || { echo "FAIL bad: $R"; exit 1; }
R=$(python3 "$PY" "$UDS_PATH" "VERIFY "); [ "$R" = "ERR BAD_REQUEST" ] || { echo "FAIL empty: $R"; exit 1; }
R=$(python3 "$PY" "$UDS_PATH" "HELLO"); [ "$R" = "ERR BAD_REQUEST" ] || { echo "FAIL np: $R"; exit 1; }
EXP=$(JWT_SECRET=$JWT_SECRET ${AB_BIN:-$HOME/arch-bench/tools/token-gen/target/release/token-gen} --from-env --sub bob --ttl 1)
sleep 2.5
R=$(python3 "$PY" "$UDS_PATH" "VERIFY $EXP"); [ "$R" = "ERR INVALID_TOKEN" ] || { echo "FAIL exp: $R"; exit 1; }
echo "PASS: c-hex-uds all tests green"
