// Composition root for rust-layered-http.
use std::env;
use std::sync::Arc;
use tiny_http::Server;

use rust_layered_http::handler::HttpHandler;
use rust_layered_http::repo::JwtRepository;
use rust_layered_http::service::{TokenVerifierService, VerifierService};

fn main() {
    let secret = env::var("JWT_SECRET").unwrap_or_default();
    if secret.is_empty() {
        eprintln!("JWT_SECRET required");
        std::process::exit(2);
    }
    let addr = env::var("LISTEN_ADDR").unwrap_or_else(|_| "127.0.0.1:8091".into());

    // Bottom-up wiring: repo → service → handler.
    let repo = JwtRepository::new(secret.as_bytes());
    let svc: Arc<dyn TokenVerifierService> = Arc::new(VerifierService::new(repo));
    // The handler holds only an Arc<dyn TokenVerifierService>, which
    // is cheap to clone — so we can pass it to every worker thread.
    let handler = Arc::new(HttpHandler::new(svc));

    let server = Server::http(&addr).expect("bind");
    eprintln!("rust-layered-http listening on {}", addr);
    for req in server.incoming_requests() {
        let h = Arc::clone(&handler);
        std::thread::spawn(move || h.handle(req));
    }
}
