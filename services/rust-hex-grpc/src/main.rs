use std::env;
use std::sync::Arc;
use tonic::transport::Server;
use rust_hex_grpc::pb::verifier_server::VerifierServer;
use rust_hex_grpc::adapter::jwt::JwtVerifier;
use rust_hex_grpc::domain::TokenVerifier;
use rust_hex_grpc::transport::GrpcAdapter;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let secret = env::var("JWT_SECRET").unwrap_or_default();
    if secret.is_empty() {
        eprintln!("JWT_SECRET required");
        std::process::exit(2);
    }
    let addr = env::var("GRPC_ADDR").unwrap_or_else(|_| "127.0.0.1:51053".into());
    let v: Arc<dyn TokenVerifier> = Arc::new(JwtVerifier::new(secret.as_bytes()));
    eprintln!("rust-hex-grpc listening on {}", addr);
    Server::builder().add_service(VerifierServer::new(GrpcAdapter::new(v))).serve(addr.parse()?).await?;
    Ok(())
}
