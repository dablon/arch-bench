# arch-bench — Build report

## Status

**24/24 services built. 24/24 services unit-tested. 24/24 mini-benchmarked.**

|  | flat | layered | hex |
|---|---|---|---|
| **Go HTTP** | ✅ | ✅ | ✅ |
| **Go UDS** | ✅ | ✅ | ✅ |
| **Go gRPC** | ✅ | ✅ | ✅ |
| **Rust HTTP** | ✅ | ✅ | ✅ |
| **Rust UDS** | ✅ | ✅ | ✅ |
| **Rust gRPC** | ✅ | ✅ | ✅ |
| **C HTTP** | ✅ | ✅ | ✅ |
| **C UDS** | ✅ | ✅ | ✅ |

(C has no gRPC column. The bench matrix is 3 transports × 3 langs × 3 archs ×
{(go,rust)×3 + c×2} = 24 services.)

## Quality gates

- **Unit tests:** every service has its own `*_test.go` / `mod tests` / `tests/test.sh`.
  All green. C services spawn the binary and curl/python the real socket.
- **E2E flows:** every service has 6–7 test cases covering ok / bad-token /
  empty / bad-body / bad-method / wrong-architecture / expired.
- **Real services, no mocks:** the JWT verify path is the real `jsonwebtoken`
  (Go/Rust) or `HMAC(EVP_sha256(), ...)` (C) call.
- **No hardcoded secrets:** all tests read `JWT_SECRET` from env. The
  `.env.example` has only placeholder values.
- **No commits of `.env`:** `.gitignore` covers it.

## Pre-push security audit (run before push)

- ✅ `gitleaks detect` not installed → manual scan: no `password=`, `api_key=`,
  `token=`, `BEGIN PRIVATE KEY` in any source file.
- ✅ `.gitignore` covers `.env`, `*.pem`, `*.key`, `target/`, `bin/`, `libssl-sym/`.
- ✅ `.env.example` is placeholder-only.
- ✅ No test fixtures contain real credentials.
- ✅ No private IPs or hostnames hardcoded.
- ✅ `git diff` clean — only source, tests, evidence, no generated noise.

## Reproducibility

```sh
# Install
go version         # 1.22+
cargo --version    # 1.74+
gcc --version      # 11+
protoc --version   # 25+
python3 --version  # 3.10+

# Build everything
(cd tools/token-gen && cargo build --release)
for s in services/go-*-{http,uds,grpc}; do (cd "$s" && go build ./... && go test -race ./...); done
for s in services/rust-*-{http,uds,grpc}; do (cd "$s" && cargo test --release); done
for s in services/c-{flat,layered,hex}-{http,uds}; do (cd "$s" && make test); done

# Bench
(cd bench/scripts/grpc-load && go build -o /tmp/grpc-load .)
DURATION=10 bash bench/scripts/run-mini.sh
python3 bench/scripts/aggregate.py evidence/mini > evidence/comparison.md
```

## Known issues / out of scope

- 100% coverage not achieved (services cover the happy path + 5–7 error paths
  per cell, ~80–90% line coverage). Mutation testing not run.
- Tier 1 (multi-hour, multi-rep) not in this commit. Plan: 5min × 3 reps ×
  24 services with cgroup sampler, in a future PR.
- k6 not used (k6 not installed in the build env). Bench uses curl + Python
  + a tiny Go gRPC client.
- The UDS bench keeps a single connection open for the run duration, which is
  realistic for an IPC verifier service but doesn't model reconnect cost.
