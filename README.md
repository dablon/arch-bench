# arch-bench

A 24-service benchmark of HMAC-JWT verification under different architectural
styles, transports, and language runtimes. The goal: produce honest, real
numbers (RPS, p50/p95/p99, throughput/MB, CPU, memory) per cell, run from a
single host with no mocks, no in-memory tricks, and a real cgroup sampler.

## Cell matrix (24 services)

|        | flat            | layered         | hex (ports/adapters) |
|--------|------------------|------------------|----------------------|
| **Go HTTP**     | go-flat-http    | go-layered-http | go-hex-http    |
| **Go UDS**      | go-flat-uds     | go-layered-uds  | go-hex-uds     |
| **Go gRPC**     | go-flat-grpc    | go-layered-grpc | go-hex-grpc    |
| **Rust HTTP**   | rust-flat-http  | rust-layered-http | rust-hex-http |
| **Rust UDS**    | rust-flat-uds   | rust-layered-uds | rust-hex-uds  |
| **Rust gRPC**   | rust-flat-grpc  | rust-layered-grpc | rust-hex-grpc |
| **C HTTP**      | c-flat-http     | c-layered-http  | c-hex-http     |
| **C UDS**       | c-flat-uds      | c-layered-uds   | c-hex-uds      |

**Transports × Languages × Architectures = 24.** Every service speaks the same
wire protocol per transport (HTTP JSON / UDS line / gRPC `verifier.v1`) and
returns the same JSON/line/proto result.

## Protocol (all 3 transports)

Request:
- **HTTP** `POST /verify` with JSON body `{"token":"<jwt>"}`
- **UDS** line `VERIFY <jwt>\n`
- **gRPC** `Verifier.Verify(VerifyRequest{token})`

Response:
- **HTTP** `200 {"valid":true,"subject":"alice","code":"OK"}` on success
  - `400 {"valid":false,"code":"ERR_BAD_REQUEST"}` for empty/missing token
  - `401 {"valid":false,"code":"ERR_INVALID_TOKEN"}` for invalid signature
  - `405` for non-POST
- **UDS** `OK <subject>\n` on success, `ERR BAD_REQUEST\n`, `ERR INVALID_TOKEN\n`
- **gRPC** `VerifyResponse{valid, subject, code}`

## Architecture styles

- **flat** — single file, no abstractions. All logic in `main()`.
- **layered** — `domain/` (verifier logic) + `transport/` (HTTP/UDS/gRPC).
  Each layer knows only about the layer below it.
- **hex (ports & adapters)** — `domain/` defines ports (`TokenVerifier`),
  `adapter/jwt/` implements the port, `adapter/{http,uds,grpc}/` consumes it.
  Domain knows nothing about transport or crypto implementation.

## Build & test (one host, no Docker required)

### Toolchain needed
- `go 1.22+`
- `rustc/cargo 1.74+`
- `gcc` + `libssl` (Ubuntu: `apt install -y libssl-dev`)
- `protoc 25+` (Ubuntu: `apt install -y protobuf-compiler`)
- `protoc-gen-go`, `protoc-gen-go-grpc` (`go install ...@latest`)
- `python3` (for the UDS test client)

### Build everything

```sh
# Token generator (used by tests + bench)
cd tools/token-gen && cargo build --release

# All Go services
for s in services/go-*-{http,uds,grpc}; do (cd "$s" && go build ./... && go test -race ./...); done

# All Rust services
for s in services/rust-*-{http,uds,grpc}; do (cd "$s" && cargo test --release); done

# All C services
for s in services/c-{flat,layered,hex}-{http,uds}; do (cd "$s" && make test); done
```

### Run a service

```sh
export JWT_SECRET="test"
export LISTEN_ADDR="127.0.0.1:8080"   # HTTP
export UDS_PATH="/tmp/foo.sock"        # UDS
export GRPC_ADDR="127.0.0.1:50051"     # gRPC

# Then:
./bin/go-flat-http            # or rust-flat-http / c-flat-http
./bin/go-flat-uds             # etc.
./bin/go-flat-grpc            # etc.
```

## Bench

`bench/scripts/`:
- `load-http.sh` — curl-based HTTP load
- `load-uds.sh` — Python UDS load client
- `load-grpc.sh` — Go gRPC load client
- `summarize.py` — turns a raw run into per-service summary
- `aggregate.py` — joins summaries into a single table
- `run-mini.sh` — 20s × 50 VU mini-bench on all 24
- `run-tier.sh` — full multi-hour Tier 1 (planned, not in this repo yet)

## Evidence

`evidence/` holds the run summaries and the comparison report.
The headline result is `evidence/comparison.md`.

## Repo privacy

This is a **private** repository. No public visibility, no public link.
