module github.com/dablon/arch-bench/services/go-hex-grpc

go 1.22

require (
	github.com/dablon/arch-bench/proto v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	google.golang.org/grpc v1.69.0
)

require (
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)

replace github.com/dablon/arch-bench/proto => ../../proto
