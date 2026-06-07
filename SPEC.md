# arch-bench Specification

## Overview
24 services: C/Rust/Go x Flat/Layered/Hex x HTTP/UDS/gRPC

## Domain: JWT Token Verifier
POST /verify - verify HS256 JWT, return valid+subject

## Services
- C: 6 services (Flat/Layered/Hex x HTTP/UDS)
- Go: 9 services (+ gRPC variants)
- Rust: 9 services (+ gRPC variants)

## Benchmark Tiers
1. Mini: 10s warmup + 10s run (smoke)
2. Short: 1min warmup + 5min run (k6)
3. Extended: 5min warmup + 30min run

## Acceptance
1. All 24 build OK
2. GET /health returns 200
3. POST /verify works
4. Mini benchmark data exists
5. evidence/comparison.md populated
