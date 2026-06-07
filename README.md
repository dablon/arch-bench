# arch-bench

Benchmark: 24 services across C/Rust/Go architectures.

## Quick Start

```bash
make all
./bench/scripts/run-mini.sh
cat evidence/comparison.md
```

## Services

C: 6 services (Flat/Layered/Hex x HTTP/UDS)
Go: 9 services (+ gRPC)
Rust: 9 services (+ gRPC)

## Testing

go test ./services/go-*/...
cargo test --manifest-path services/rust-*/Cargo.toml

## Results

See evidence/comparison.md
