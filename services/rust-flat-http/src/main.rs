// rust-flat-http: FLAT-architecture HMAC-SHA256 JWT verifier over HTTP.
//
// Architectural style: FLAT. One file, one function per concern, no
// abstractions. The wire format, the JWT logic, the HTTP routing, the
// JSON encoding, the env config are all visible in ~150 lines.
use jsonwebtoken::{decode, DecodingKey, Validation};
use serde::{Deserialize, Serialize};
use std::env;
use std::io::Read;
use std::sync::Arc;
use tiny_http::{Header, Method, Response, Server};

#[derive(Deserialize)]
struct VerifyRequest { token: String }

#[derive(Serialize)]
struct VerifyResponse {
    valid: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    subject: Option<String>,
    code: String,
}

fn verify_token(token: &str, secret: &[u8]) -> VerifyResponse {
    if token.is_empty() {
        return VerifyResponse { valid: false, subject: None, code: "ERR_BAD_REQUEST".into() };
    }
    let v = Validation::new(jsonwebtoken::Algorithm::HS256);
    match decode::<serde_json::Value>(token, &DecodingKey::from_secret(secret), &v) {
        Ok(data) => {
            let sub = data.claims.get("sub").and_then(|v| v.as_str()).unwrap_or("").to_string();
            VerifyResponse { valid: true, subject: Some(sub), code: "OK".into() }
        }
        Err(_) => VerifyResponse { valid: false, subject: None, code: "ERR_INVALID_TOKEN".into() },
    }
}

fn handle(mut req: tiny_http::Request, secret: Arc<Vec<u8>>) {
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
    let response = verify_token(&token, &secret);
    let status = if response.valid { 200 } else if response.code == "ERR_BAD_REQUEST" { 400 } else { 401 };
    let _ = req.respond(Response::from_string(serde_json::to_string(&response).unwrap())
        .with_status_code(status)
        .with_header(Header::from_bytes(&b"Content-Type"[..], &b"application/json"[..]).unwrap()));
}

fn main() {
    let secret = env::var("JWT_SECRET").unwrap_or_default();
    if secret.is_empty() {
        eprintln!("JWT_SECRET required");
        std::process::exit(2);
    }
    let addr = env::var("LISTEN_ADDR").unwrap_or_else(|_| "127.0.0.1:8090".into());
    let secret = Arc::new(secret.into_bytes());
    let server = Server::http(&addr).expect("bind");
    eprintln!("rust-flat-http listening on {}", addr);
    for req in server.incoming_requests() {
        let s = secret.clone();
        std::thread::spawn(move || handle(req, s));
    }
}
