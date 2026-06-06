# arch-bench — Comparison report

**Mini-bench:** 10s × 1 worker (single-threaded sequential) per cell.
**Host:** single Linux box, no Docker, services ran as bare processes.
**Load profile:** 1 valid JWT per run, sent sequentially, latency recorded per request.

> ⚠️ This is a single-writer mini-bench. Absolute RPS numbers reflect a single
> connection in a single process. The relative ordering within each transport
> is the signal.

## Numbers (sorted by transport, then cell)

| cell | total | p50 (ms) | p95 (ms) | p99 (ms) | max (ms) | RPS |
|------|-------|----------|----------|----------|----------|-----|
| go-flat-http        | 488  | 15 | 19 | 22 | 28 | 49    |
| go-layered-http     | 525  | 14 | 18 | 20 | 32 | 52    |
| go-hex-http         | 530  | 14 | 18 | 20 | 22 | 53    |
| rust-flat-http      | 503  | 15 | 18 | 19 | 29 | 50    |
| rust-layered-http   | 505  | 14 | 19 | 25 | 47 | 50    |
| rust-hex-http       | 529  | 14 | 18 | 19 | 21 | 53    |
| c-flat-http         | 437  | 18 | 22 | 24 | 26 | 44    |
| c-layered-http      | 434  | 18 | 22 | 25 | 26 | 43    |
| c-hex-http          | 449  | 17 | 21 | 23 | 29 | 45    |
| go-flat-uds         | 123426 | 0.07 | 0.11 | 0.16 | 12.33 | 12343 |
| go-layered-uds      | 124876 | 0.07 | 0.11 | 0.15 | 4.45  | 12488 |
| go-hex-uds          | 124464 | 0.07 | 0.11 | 0.15 | 6.08  | 12446 |
| rust-flat-uds       | 192094 | 0.04 | 0.07 | 0.09 | 4.96  | 19209 |
| rust-layered-uds    | 175544 | 0.05 | 0.07 | 0.10 | 3.73  | 17554 |
| rust-hex-uds        | 174215 | 0.05 | 0.07 | 0.10 | 6.01  | 17422 |
| c-flat-uds          | 142011 | 0.06 | 0.09 | 0.16 | 11.27 | 14201 |
| c-layered-uds       | 163438 | 0.05 | 0.08 | 0.10 | 4.15  | 16344 |
| c-hex-uds           | 153311 | 0.06 | 0.09 | 0.11 | 7.70  | 15331 |
| go-flat-grpc        | 24915 | 0.35 | 0.56 | 0.84 | 4.42  | 2492  |
| go-layered-grpc     | 23574 | 0.36 | 0.62 | 1.14 | 10.60 | 2357  |
| go-hex-grpc         | 24517 | 0.36 | 0.60 | 0.95 | 8.98  | 2452  |
| rust-flat-grpc      | 32881 | 0.27 | 0.39 | 0.59 | 11.31 | 3288  |
| rust-layered-grpc   | 32424 | 0.28 | 0.40 | 0.60 | 3.95  | 3242  |
| rust-hex-grpc       | 34268 | 0.26 | 0.38 | 0.54 | 4.87  | 3427  |

## What you can read off this

**Transport matters more than architecture or language.**

- UDS: ~12k–19k RPS, sub-ms p50.
- gRPC: ~2.3k–3.4k RPS, sub-ms p50 but with extra ~0.3ms of HTTP/2 framing.
- HTTP (curl over loopback): ~43–53 RPS, ~14–18ms p50. The HTTP numbers are
  bottlenecked by `curl -s` per-request fork/exec, not by the server.

**Within a transport, the architecture deltas are tiny.**

- HTTP: layered/hex is ~1–2 RPS faster than flat (within curl noise).
- UDS: rust-flat is faster than rust-layered/hex by ~1.5k RPS (8%). Likely
  because the flat path doesn't do extra struct/closure indirection.
- gRPC: hex is fastest by ~1–2% (within noise).
- C HTTP: hex/flat are 1 RPS faster than layered (within noise).

**Within a transport, the language deltas:**

- UDS: rust ≈ c > go. Go's net package adds ~30µs per round-trip vs Rust/C
  talking raw `UnixStream` / `UnixListener`.
- gRPC: rust > go by ~30% (Rust tonic 0.13 vs Go grpc 1.69).
- HTTP: tied — the bottleneck is curl, not the server.

## Caveats

1. **Single worker.** RPS scales with concurrency. Run a Tier 1 with
   50/100/200 concurrent connections to get absolute numbers.
2. **No warm-up loop.** First request of a run eats process start + accept.
3. **Loopback only.** Network isn't the bottleneck; the syscall stack is.
4. **No CPU/mem telemetry here.** Tier 1 will add cgroup sampling.
5. **Single host.** A bigger box will raise the absolute ceiling, but the
   relative ordering within a transport is stable.

## How to reproduce

```sh
cd /workspace/projects/arch-bench
DURATION=10 bash bench/scripts/run-mini.sh
python3 bench/scripts/aggregate.py evidence/mini > evidence/comparison.md
```
