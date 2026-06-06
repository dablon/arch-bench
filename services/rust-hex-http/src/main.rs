// IN adapter: HTTP
use rust_hex_http::adapter::jwt::JwtVerifier;
use rust_hex_http::domain::TokenVerifier;
use serde::Deserialize;
use std::env;
use std::sync::Arc;
use tiny_http::{Header, Method, Response, Server};

fn handle(mut req: tiny_http::Request, verifier: Arc<dyn TokenVerifier>) {
    let url = req.url().to_string();
    let method = req.method().clone();
    if url == "/health" {
        let _ = req.respond(Response::from_string(r#"{"status":"ok"}"#).with_header(
            Header::from_bytes(&b"Content-Type"[..], &b"application/json"[..]).unwrap(),
        ));
        return;
    }
    if url != "/verify" {
        let _ = req.respond(Response::from_string("not found").with_status_code(404));
        return;
    }
    if method != Method::Post {
        let _ = req.respond(Response::from_string("method not allowed").with_status_code(405));
        return;
    }
    let mut body = String::new();
    use std::io::Read;
    let _ = req.as_reader().read_to_string(&mut body);
    #[derive(Deserialize)]
    struct VerifyRequest { token: String }
    let parsed: Result<VerifyRequest, _> = serde_json::from_str(&body);
    let token = match parsed {
        Ok(p) if !p.token.is_empty() => p.token,
        _ => String::new(),
    };
    let response = verifier.verify(&token);
    let status = if response.valid { 200 } else if response.code == "ERR_BAD_REQUEST" { 400 } else { 401 };
    let _ = req.respond(
        Response::from_string(serde_json::to_string(&response).unwrap())
            .with_status_code(status)
            .with_header(Header::from_bytes(&b"Content-Type"[..], &b"application/json"[..]).unwrap()),
    );
}

fn main() {
    let secret = env::var("JWT_SECRET").unwrap_or_default();
    if secret.is_empty() {
        eprintln!("JWT_SECRET required");
        std::process::exit(2);
    }
    let addr = env::var("LISTEN_ADDR").unwrap_or_else(|_| "127.0.0.1:8092".into());
    let verifier: Arc<dyn TokenVerifier> = Arc::new(JwtVerifier::new(secret.as_bytes()));
    let server = Server::http(&addr).expect("bind");
    eprintln!("rust-hex-http listening on {}", addr);
    for req in server.incoming_requests() {
        let v = verifier.clone();
        std::thread::spawn(move || handle(req, v));
    }
}
