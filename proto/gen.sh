#!/bin/sh
set -e
protoc -I. --go_out=. --go_opt=module=github.com/dablon/arch-bench/proto --go-grpc_out=. --go-grpc_opt=module=github.com/dablon/arch-bench/proto verifier.proto
