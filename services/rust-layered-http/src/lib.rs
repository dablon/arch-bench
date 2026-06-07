// rust-layered-http: LAYERED-architecture HMAC-SHA256 JWT verifier.
//
// Modules (imports go strictly downward):
//   repo     — data access: the JWT verifier, no business knowledge
//   service  — business logic: maps repo errors to domain result codes
//   handler  — HTTP transport: parses request, calls service, encodes response
//
// In Rust, modules live in submodules of the crate root. Each module is
// its own file under src/.

pub mod repo;
pub mod service;
pub mod handler;

pub use repo::JwtRepository;
pub use service::{VerifyOutcome, ResultCode, TokenVerifierService};
pub use handler::HttpHandler;
