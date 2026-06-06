# arch-bench — Specification

## What this is

A controlled micro-benchmark that measures the cost of HMAC-SHA256 JWT
verification across 3 axes:

1. **Transport** (HTTP, UDS, gRPC)
2. **Language** (Go, Rust, C)
3. **Architecture** (flat, layered, hex/ports-adapters)

Result: a 24-row table of (req/s, p50, p95, p99, MB/s) on a fixed host,
captured with a fixed load profile.

## Why

The "right" architecture for a verifier service is a debate that often turns
on vibes. This repo replaces the vibes with numbers. It's not the final
answer — the absolute numbers depend on the host and the load profile — but
the **relative** numbers ("gRPC is N× faster than HTTP in this loop", "Go
beats Rust by X% in microseconds", "the hex architecture adds Y ns/req")
are stable enough to inform decisions.

## Non-goals

- Multi-host benchmarks (single host, single process tree)
- Persistent state (no DB, no cache, no rate limiting)
- Auth above the verifier itself (we trust the `JWT_SECRET` env var)
- Public visibility (this repo is private)
- The "right" architecture (we report, you decide)

## Wire protocol

All three transports speak the same logical request/response:

| Direction | Shape |
|---|---|
| Request | A JWT string |
| Response OK | `valid=true, subject=<sub>, code="OK"` |
| Response bad-request | `code="ERR_BAD_REQUEST"` (empty or missing token) |
| Response invalid | `code="ERR_INVALID_TOKEN"` (signature, alg, expiry fail) |

The error codes are stable across transports so a test can assert on them
through HTTP, UDS, or gRPC.

## How cells are built

Each cell is a single-purpose service that:
1. Reads `JWT_SECRET` from env (the only config).
2. Listens on its transport-specific address (also from env, with sensible defaults).
3. Verifies incoming tokens.
4. Returns the canonical response.

The three architecture styles in each cell vary only how the verifier logic is
packaged:
- `flat` — one source file, inline.
- `layered` — domain package + transport package, wired in main.
- `hex` — domain port + jwt adapter + transport adapter, wired in main.

The functional behavior is identical. The benchmarks measure the cost of the
extra layers.

## Methodology

### Mini bench (in this repo)

- Duration: 20s
- Concurrency: 50 VU
- Token: freshly minted per VU per request? No — **one token per VU**, reused
  for the full 20s. This isolates the verifier cost from the generator cost.
- Sample every request's latency.
- 1 repetition per cell.

The mini bench runs in ~10 min total on a single host. Numbers are good for
relative comparison; absolute RPS depends on the host.

### Tier 1 (planned, not yet in repo)

- Duration: 5 min × 3 reps per cell
- Concurrency: 50, 100, 200 VU
- Cold-cache control: 1 min idle between reps
- Telemetry: cgroup CPU/mem/throttling, ctx switches, fds, net
- Output: `evidence/tier1/<cell>.{rps,p99,cpu,mem}.json` per rep

Tier 1 takes ~6 hours on a single OVH2 box. Will be added in a future commit.

## Status (this commit)

All 24 services built and unit-tested. Mini-bench report in
`evidence/comparison.md` with 24 rows. Tier 1 deferred.
