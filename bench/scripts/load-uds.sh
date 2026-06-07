#!/bin/sh
# UDS load generator (Python, sequential). Emits one latency per line to stdout
# in milliseconds. The last line is a "# summary ..." comment.
# Usage: load-uds.sh <socket> <token> <duration_sec>

set -e
SOCK="$1"; TOKEN="$2"; DURATION="$3"
python3 - << EOF
import socket, time, sys
sock_path = "$SOCK"
token = "$TOKEN"
duration = $DURATION
s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
s.connect(sock_path)
s.settimeout(5)
end = time.time() + duration
ok = bad = br = 0
total = 0
while time.time() < end:
    t0 = time.time()
    s.sendall(("VERIFY " + token + "\\n").encode())
    buf = b""
    while not buf.endswith(b"\\n"):
        chunk = s.recv(4096)
        if not chunk: break
        buf += chunk
    t1 = time.time()
    line = buf.decode().rstrip()
    if line.startswith("OK "): ok += 1
    elif line.startswith("ERR INVALID_TOKEN"): bad += 1
    elif line.startswith("ERR BAD_REQUEST"): br += 1
    total += 1
    print(f"{(t1-t0)*1000:.3f}")
print(f"# summary ok={ok} bad={bad} br={br} total={total}", file=sys.stderr)
EOF
