// rust-layered-grpc: LAYERED-architecture HMAC-SHA256 JWT verifier over gRPC.
//
// Modules (imports go strictly downward):
//   repo     — JWT data access
//   service  — business logic
//   handler  — gRPC transport (proto types)
pub mod pb {
    tonic::include_proto!("verifier");
}
pub mod repo;
pub mod service;
pub mod handler;
