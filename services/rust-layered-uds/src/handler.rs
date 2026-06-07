// UDS line protocol handler. Reads "VERIFY <jwt>\n", writes "OK <sub>\n"
// or "ERR <code>\n". The transport knows nothing about JWT.
use crate::service::{ResultCode, TokenVerifierService};
use std::io::{BufRead, BufReader, Write};
use std::os::unix::net::UnixStream;
use std::sync::Arc;

pub fn handle(mut stream: UnixStream, svc: Arc<dyn TokenVerifierService>) {
    let mut reader = BufReader::new(stream.try_clone().unwrap());
    let mut buf = String::new();
    loop {
        buf.clear();
        let n = match reader.read_line(&mut buf) { Ok(n) => n, Err(_) => break };
        if n == 0 { break; }
        let trimmed = buf.trim_end_matches(|c| c == '\n' || c == '\r');
        let resp: String = match trimmed.strip_prefix("VERIFY ") {
            Some(tok) => {
                let r = svc.verify(tok);
                if r.valid {
                    format!("OK {}\n", r.subject)
                } else if r.code == ResultCode::BadRequest {
                    "ERR BAD_REQUEST\n".to_string()
                } else {
                    "ERR INVALID_TOKEN\n".to_string()
                }
            }
            None => "ERR BAD_REQUEST\n".to_string(),
        };
        let _ = stream.write_all(resp.as_bytes());
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::service::{ResultCode as RC, VerifyOutcome};
    struct Stub(VerifyOutcome);
    impl TokenVerifierService for Stub { fn verify(&self, _: &str) -> VerifyOutcome { self.0.clone() } }
    #[test] fn smoke() {
        let s: Arc<dyn TokenVerifierService> = Arc::new(Stub(VerifyOutcome { valid: true, subject: "alice".into(), code: RC::Ok }));
        let r = s.verify("x");
        assert!(r.valid);
    }
}
