#!/usr/bin/env python3
"""Aggregate per-cell CSVs into one big table (RPS, p50, p95, p99, max)."""
import os, sys, statistics

DUR = 10  # seconds (matches run-mini default)
DIR = sys.argv[1] if len(sys.argv) > 1 else "evidence/mini"

cells = []
for f in sorted(os.listdir(DIR)):
    if not f.endswith(".csv"):
        continue
    name = f[:-4]
    path = os.path.join(DIR, f)
    lats = []
    ok = bad = br = 0
    with open(path) as fh:
        for line in fh:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "," in line:
                p = line.split(",")
                try:
                    code = int(p[0])
                    if code == 200: ok += 1
                    elif code == 401: bad += 1
                    elif code == 400: br += 1
                    lats.append(float(p[1]))
                except (ValueError, IndexError):
                    pass
            else:
                try:
                    lats.append(float(line))
                except ValueError:
                    pass
    if not lats:
        cells.append((name, 0, 0, 0, 0, 0, 0))
        continue
    lats.sort()
    n = len(lats)
    def p(q):
        i = int(q * n)
        if i >= n: i = n - 1
        return lats[i]
    total = ok + bad + br
    if total == 0: total = n
    rps = total / DUR
    cells.append((name, total, p(0.5), p(0.95), p(0.99), lats[-1], rps))

# Print markdown table
print("| cell | total | p50 (ms) | p95 (ms) | p99 (ms) | max (ms) | RPS |")
print("|------|-------|----------|----------|----------|----------|-----|")
for c in cells:
    name, total, p50, p95, p99, mx, rps = c
    print(f"| {name} | {total} | {p50:.2f} | {p95:.2f} | {p99:.2f} | {mx:.2f} | {rps:.0f} |")
