#!/usr/bin/env python3
"""Automated benchmark runner for arch-bench"""
import json, time
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
RESULTS_DIR = Path("/home/nalcaraz/arch-bench/evidence/benchmark-results")
SERVICES = {
  "rust-flat-http": {"port": 8090},
  "rust-hex-http": {"port": 8091},
  "rust-layered-http": {"port": 8092},
  "go-flat-http": {"port": 8080},
  "go-hex-http": {"port": 8081},
}
def run_http_benchmark(name, port, duration=30, workers=100):
  import urllib.request
  token = open("/tmp/token.txt").read().strip()
  results = []; ok = errors = 0
  start = time.time(); end = start + duration
  def do_req():
    nonlocal ok, errors
    try:
      t0 = time.time()
      req = urllib.request.Request(f"http://127.0.0.1:{port}/verify",
        data=json.dumps({"token": token}).encode(),
        headers={"Content-Type": "application/json"}, method="POST")
      resp = urllib.request.urlopen(req, timeout=5)
      t1 = time.time()
      if resp.status == 200: ok += 1; return (t1 - t0) * 1000
    except: errors += 1; return 0
  with ThreadPoolExecutor(max_workers=workers) as ex:
    futs = []
    while time.time() < end: futs.append(ex.submit(do_req))
    for f in futs:
      lat = f.result(); results.append(lat) if lat > 0 else None
  results.sort(); n = len(results); dur = time.time() - start
  tps = n / dur if dur > 0 else 0
  return {"name": name, "protocol": "http", "tps": round(tps, 1),
    "avg_ms": round(sum(results)/n, 2) if n > 0 else 0,
    "p50": round(results[int(n*0.50)], 2) if n > 0 else 0,
    "p95": round(results[int(n*0.95)], 2) if n > 0 else 0,
    "p99": round(results[int(n*0.99)], 2) if n > 0 else 0,
    "max": round(results[-1], 2) if results else 0, "ok": ok}
def main():
  timestamp = time.strftime("%Y-%m-%d_%H-%M-%S")
  results = {"timestamp": timestamp, "services": [], "meta": {"duration": 30, "workers": 100}}
  RESULTS_DIR.mkdir(parents=True, exist_ok=True)
  for name, cfg in SERVICES.items():
    print(f"[BM] Testing {name}...")
    result = run_http_benchmark(name, cfg["port"])
    results["services"].append(result)
    tps_val = result["tps"]; print("[BM] " + name + ": " + str(tps_val) + " TPS")
  out_file = RESULTS_DIR / f"results_{timestamp}.json"
  with open(out_file, "w") as f: json.dump(results, f, indent=2)
  print(f"[BM] Saved: {out_file}")
if __name__ == "__main__": main()
