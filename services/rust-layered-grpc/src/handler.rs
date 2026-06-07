// handler: gRPC transport. Knows proto types and the service interface
// and nothing else.
use crate::pb::{verifier_server::Verifier, VerifyRequest, VerifyResponse};
use crate::service::{ResultCode, TokenVerifierService};
use std::sync::Arc;
use tonic::{Request, Response, Status};

pub struct GrpcHandler {
    svc: Arc<dyn TokenVerifierService>,
}

impl GrpcHandler {
    pub fn new(svc: Arc<dyn TokenVerifierService>) -> Self { Self { svc } }
}

#[tonic::async_trait]
impl Verifier for GrpcHandler {
    async fn verify(&self, req: Request<VerifyRequest>) -> Result<Response<VerifyResponse>, Status> {
        let r = self.svc.verify(&req.into_inner().token);
        let code = if r.valid {
            ResultCode::Ok
        } else if r.code == ResultCode::BadRequest {
            ResultCode::BadRequest
        } else {
            ResultCode::InvalidToken
        };
        Ok(Response::new(VerifyResponse {
            valid: r.valid,
            subject: r.subject,
            code: code.as_str().to_string(),
        }))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::service::{VerifyOutcome, ResultCode as RC};

    struct Stub(VerifyOutcome);
    impl TokenVerifierService for Stub {
        fn verify(&self, _: &str) -> VerifyOutcome { self.0.clone() }
    }

    #[tokio::test]
    async fn ok() {
        let svc: Arc<dyn TokenVerifierService> = Arc::new(Stub(VerifyOutcome { valid: true, subject: "alice".into(), code: RC::Ok }));
        let h = GrpcHandler::new(svc);
        let r = h.verify(Request::new(VerifyRequest { token: "x".into() })).await.unwrap().into_inner();
        assert!(r.valid);
        assert_eq!(r.code, "OK");
    }

    #[tokio::test]
    async fn invalid() {
        let svc: Arc<dyn TokenVerifierService> = Arc::new(Stub(VerifyOutcome { valid: false, subject: "".into(), code: RC::InvalidToken }));
        let h = GrpcHandler::new(svc);
        let r = h.verify(Request::new(VerifyRequest { token: "x".into() })).await.unwrap().into_inner();
        assert_eq!(r.code, "ERR_INVALID_TOKEN");
    }

    #[tokio::test]
    async fn bad_request() {
        let svc: Arc<dyn TokenVerifierService> = Arc::new(Stub(VerifyOutcome { valid: false, subject: "".into(), code: RC::BadRequest }));
        let h = GrpcHandler::new(svc);
        let r = h.verify(Request::new(VerifyRequest { token: "".into() })).await.unwrap().into_inner();
        assert_eq!(r.code, "ERR_BAD_REQUEST");
    }
}
