#!/usr/bin/env python3
"""Summarize a per-cell CSV/TSV: total, ok/bad, p50, p95, p99, rps."""
import sys, os, statistics

if len(sys.argv) < 2:
    print("usage: summarize.py <cell.csv>", file=sys.stderr)
    sys.exit(1)

path = sys.argv[1]
if not os.path.exists(path):
    print(f"# missing {path}", file=sys.stderr)
    sys.exit(0)

# Detect format: HTTP CSVs are "code,lat_ms" with first column 200/401/...
# UDS/gRPC are bare latencies (one per line).
lats = []
codes = {}
with open(path) as f:
    for line in f:
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        if "," in line:
            parts = line.split(",")
            if len(parts) >= 2 and parts[0].isdigit():
                code = int(parts[0])
                codes[code] = codes.get(code, 0) + 1
                try:
                    lats.append(float(parts[1]))
                except ValueError:
                    pass
        else:
            try:
                lats.append(float(line))
            except ValueError:
                pass

if not lats:
    print(f"# {path}: no data", file=sys.stderr)
    sys.exit(0)

lats.sort()
n = len(lats)
def p(q):
    if n == 0: return 0
    i = int(q * n)
    if i >= n: i = n - 1
    return lats[i]

total = sum(codes.values()) if codes else n
ok = codes.get(200, 0)
duration = int(os.environ.get("DURATION", "20"))
rps = total / duration if duration else 0

print(f"cell={os.path.basename(path)} total={total} ok={ok} "
      f"p50={p(0.5):.2f} p95={p(0.95):.2f} p99={p(0.99):.2f} "
      f"max={lats[-1]:.2f} rps={rps:.1f}")
print(f"  codes: {codes}", file=sys.stderr)
