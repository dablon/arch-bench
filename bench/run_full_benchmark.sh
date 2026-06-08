#!/bin/bash
set -e
BASE="/home/nalcaraz/arch-bench"
EVIDENCE="$BASE/evidence/benchmark-results"
log() { echo -e "\033[0;36m[BM]\033[0m $1"; }
cleanup() { log "Cleaning up..."; for svc in rust-flat-http rust-hex-http rust-layered-http go-flat-http go-hex-http go-layered-http; do pkill -f "$svc" 2>/dev/null || true; done; sleep 2; }
start_service() { local n=$1 b=$2 p=$3; log "Starting $n on $p..."; (env JWT_SECRET="***" LISTEN_ADDR="127.0.0.1:$p" $b > /tmp/$n.log 2>&1 &); sleep 3; ss -tlnp | grep -q ":$p " && log "$n OK" || log "$n FAILED"; }
cleanup
log "Starting services..."
start_service "rust-flat-http" "/home/nalcaraz/arch-bench/services/rust-flat-http/target/release/rust-flat-http" 8090
start_service "rust-hex-http" "/home/nalcaraz/arch-bench/services/rust-hex-http/target/release/rust-hex-http" 8091
start_service "rust-layered-http" "/home/nalcaraz/arch-bench/services/rust-layered-http/target/release/rust-layered-http" 8092
start_service "go-flat-http" "/home/nalcaraz/arch-bench/services/go-flat-http/go-flat-http" 8080
start_service "go-hex-http" "/home/nalcaraz/arch-bench/services/go-hex-http/go-hex-http" 8081
start_service "go-layered-http" "/home/nalcaraz/arch-bench/services/go-layered-http/go-layered-http" 8082
log "Running benchmark..."
python3 "$BASE/bench/benchmark_runner.py"
LATEST=$(ls -t "$EVIDENCE"/results_*.json 2>/dev/null | head -1)
if [ -n "$LATEST" ]; then python3 "$BASE/bench/generate_report.py" "$LATEST"; fi
cleanup
cd "$BASE"
git add -A
git commit -m "auto: benchmark $(date +%Y-%m-%d_%H-%M)"
git push origin main 2>&1 | tail -3
log "Done!"
