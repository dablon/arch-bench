// rust-hex-uds: HEXAGONAL architecture.
// domain/ has zero external deps (no jsonwebtoken, no tokio, no UDS).
// infrastructure/ implements the domain ports.
pub mod domain;
pub mod infrastructure;
