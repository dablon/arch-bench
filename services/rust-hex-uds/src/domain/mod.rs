// rust-hex-uds/src/domain: pure domain layer. NO I/O, NO network,
// NO JSON library. Only the abstract notion of "verify this token".
pub mod port;
pub mod result;
pub mod usecase;
