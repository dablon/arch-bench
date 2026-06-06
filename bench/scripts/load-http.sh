#!/bin/sh
# HTTP load generator (curl-based, sequential within a worker).
# Usage: load-http.sh <port> <token> <duration_sec> <output_csv>

set -e
PORT="$1"; TOKEN="$2"; DURATION="$3"; OUT="$4"
END=$(($(date +%s) + DURATION))
N_OK=0; N_401=0; N_400=0; N_OTHER=0
TOTAL_LAT_MS=0
COUNT=0
> "$OUT"
while [ "$(date +%s)" -lt "$END" ]; do
    T0=$(date +%s%N)
    CODE=$(curl -s -m 5 -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "{\"token\":\"$TOKEN\"}" "http://127.0.0.1:$PORT/verify")
    T1=$(date +%s%N)
    LAT_MS=$(( (T1 - T0) / 1000000 ))
    echo "$CODE,$LAT_MS" >> "$OUT"
    case "$CODE" in
        200) N_OK=$((N_OK+1));;
        401) N_401=$((N_401+1));;
        400) N_400=$((N_400+1));;
        *) N_OTHER=$((N_OTHER+1));;
    esac
    TOTAL_LAT_MS=$((TOTAL_LAT_MS + LAT_MS))
    COUNT=$((COUNT+1))
done
RPS=$(( COUNT / DURATION ))
AVG_LAT=$(( TOTAL_LAT_MS / (COUNT > 0 ? COUNT : 1) ))
echo "ok=$N_OK 401=$N_401 400=$N_400 other=$N_OTHER total=$COUNT rps=$RPS avg_ms=$AVG_LAT" >&2
