// Composition root for rust-layered-grpc.
use std::env;
use std::sync::Arc;
use tonic::transport::Server;

use rust_layered_grpc::handler::GrpcHandler;
use rust_layered_grpc::pb::verifier_server::VerifierServer;
use rust_layered_grpc::repo::JwtRepository;
use rust_layered_grpc::service::{TokenVerifierService, VerifierService};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let secret = env::var("JWT_SECRET").unwrap_or_default();
    if secret.is_empty() {
        eprintln!("JWT_SECRET required");
        std::process::exit(2);
    }
    let addr = env::var("GRPC_ADDR").unwrap_or_else(|_| "127.0.0.1:51052".into());

    let repo = JwtRepository::new(secret.as_bytes());
    let svc: Arc<dyn TokenVerifierService> = Arc::new(VerifierService::new(repo));
    let handler = GrpcHandler::new(svc);

    eprintln!("rust-layered-grpc listening on {}", addr);
    Server::builder()
        .add_service(VerifierServer::new(handler))
        .serve(addr.parse()?)
        .await?;
    Ok(())
}
