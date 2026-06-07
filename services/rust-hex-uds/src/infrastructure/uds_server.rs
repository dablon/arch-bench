// rust-hex-uds/src/infrastructure/uds_server.rs: the INCOMING adapter.
// UDS line protocol + composition root.
use crate::domain::port::TokenVerifierPort;
use crate::domain::result::{DomainCode, VerificationResult};
use crate::domain::usecase;
use crate::infrastructure::jwt_adapter::HmacJwtAdapter;
use std::io::{BufRead, BufReader, Write};
use std::os::unix::net::UnixStream;
use std::sync::Arc;

pub fn handle(mut stream: UnixStream, port: Arc<dyn TokenVerifierPort>) {
    let mut reader = BufReader::new(stream.try_clone().unwrap());
    let mut buf = String::new();
    loop {
        buf.clear();
        let n = match reader.read_line(&mut buf) { Ok(n) => n, Err(_) => break };
        if n == 0 { break; }
        let trimmed = buf.trim_end_matches(|c| c == '\n' || c == '\r');
        let resp = match trimmed.strip_prefix("VERIFY ") {
            Some(tok) => {
                let r: VerificationResult = usecase::verify(&*port, tok);
                if r.valid { format!("OK {}\n", r.subject) }
                else if r.code == DomainCode::BadRequest { "ERR BAD_REQUEST\n".to_string() }
                else { "ERR INVALID_TOKEN\n".to_string() }
            }
            None => "ERR BAD_REQUEST\n".to_string(),
        };
        let _ = stream.write_all(resp.as_bytes());
    }
}

pub fn build_port(secret: Vec<u8>) -> Arc<dyn TokenVerifierPort> {
    Arc::new(HmacJwtAdapter::new(secret))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::port::{RepoError, RepoResult};
    struct Fake(Result<String, RepoError>);
    impl TokenVerifierPort for Fake { fn verify(&self, _: &str) -> RepoResult { self.0.clone() } }
    #[test] fn smoke() { let p: Arc<dyn TokenVerifierPort> = Arc::new(Fake(Ok("a".into()))); assert_eq!(p.verify("t").unwrap(), "a"); }
}
