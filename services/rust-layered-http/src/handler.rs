// handler: HTTP transport. Knows HTTP, JSON, status codes. Doesn't
// know about JWT, claims, or crypto.
use crate::service::{ResultCode, TokenVerifierService};
use serde::Deserialize;
use std::io::Read;
use std::sync::Arc;
use tiny_http::{Header, Method, Response};

#[derive(Deserialize)]
struct VerifyRequest { token: String }

pub struct HttpHandler {
    svc: Arc<dyn TokenVerifierService>,
}

impl HttpHandler {
    pub fn new(svc: Arc<dyn TokenVerifierService>) -> Self { Self { svc } }

    pub fn handle(&self, mut req: tiny_http::Request) {
        let url = req.url().to_string();
        if url == "/health" {
            let _ = req.respond(Response::from_string(r#"{"status":"ok"}"#)
                .with_header(Header::from_bytes(&b"Content-Type"[..], &b"application/json"[..]).unwrap()));
            return;
        }
        if url != "/verify" {
            let _ = req.respond(Response::from_string("not found").with_status_code(404));
            return;
        }
        if *req.method() != Method::Post {
            let _ = req.respond(Response::from_string("method not allowed").with_status_code(405));
            return;
        }
        let mut body = String::new();
        let _ = req.as_reader().read_to_string(&mut body);
        let parsed: Result<VerifyRequest, _> = serde_json::from_str(&body);
        let token = match parsed {
            Ok(p) if !p.token.is_empty() => p.token,
            _ => String::new(),
        };
        let res = self.svc.verify(&token);
        let (status, code) = if res.valid {
            (200, ResultCode::Ok.as_str())
        } else if res.code == ResultCode::BadRequest {
            (400, ResultCode::BadRequest.as_str())
        } else {
            (401, ResultCode::InvalidToken.as_str())
        };
        let body = if res.valid {
            format!(r#"{{"valid":true,"subject":"{}","code":"OK"}}"#, res.subject)
        } else {
            format!(r#"{{"valid":false,"code":"{}"}}"#, code)
        };
        let _ = req.respond(Response::from_string(body)
            .with_status_code(status)
            .with_header(Header::from_bytes(&b"Content-Type"[..], &b"application/json"[..]).unwrap()));
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::service::{VerifyOutcome, ResultCode, TokenVerifierService};

    struct Stub(VerifyOutcome);
    impl TokenVerifierService for Stub {
        fn verify(&self, _: &str) -> VerifyOutcome { self.0.clone() }
    }

    #[test]
    fn handler_dispatches_to_service() {
        // We can't easily drive tiny_http from tests, but we can assert
        // the handler is constructed correctly.
        let svc: Arc<dyn TokenVerifierService> = Arc::new(Stub(VerifyOutcome { valid: true, subject: "alice".into(), code: ResultCode::Ok }));
        let _h = HttpHandler::new(svc);
    }
}
