use std::env;
use std::sync::Arc;
use tonic::transport::Server;
use rust_flat_grpc::pb::{verifier_server::{Verifier, VerifierServer}, VerifyRequest, VerifyResponse};
use rust_flat_grpc::verify_token;

pub struct VerifierSvc { secret: Arc<Vec<u8>> }

#[tonic::async_trait]
impl Verifier for VerifierSvc {
    async fn verify(&self, _req: tonic::Request<VerifyRequest>) -> tonic::Result<tonic::Response<VerifyResponse>> {
        let r = verify_token(&_req.into_inner().token, &self.secret);
        Ok(tonic::Response::new(r))
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let secret = env::var("JWT_SECRET").unwrap_or_default();
    if secret.is_empty() {
        eprintln!("JWT_SECRET required");
        std::process::exit(2);
    }
    let addr = env::var("GRPC_ADDR").unwrap_or_else(|_| "127.0.0.1:51051".into());
    let secret = Arc::new(secret.into_bytes());
    let svc = VerifierSvc { secret };
    eprintln!("rust-flat-grpc listening on {}", addr);
    Server::builder()
        .add_service(VerifierServer::new(svc))
        .serve(addr.parse()?)
        .await?;
    Ok(())
}
