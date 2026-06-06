// Generated proto types live alongside src
pub mod pb {
    tonic::include_proto!("verifier");
}

use jsonwebtoken::{decode, DecodingKey, Validation};
use pb::{VerifyRequest, VerifyResponse};

pub fn verify_token(token: &str, secret: &[u8]) -> VerifyResponse {
    if token.is_empty() {
        return VerifyResponse { valid: false, subject: "".into(), code: "ERR_BAD_REQUEST".into() };
    }
    let mut v = Validation::new(jsonwebtoken::Algorithm::HS256);
    v.leeway = 0;
    match decode::<serde_json::Value>(token, &DecodingKey::from_secret(secret), &v) {
        Ok(data) => {
            let sub = data.claims.get("sub").and_then(|v| v.as_str()).unwrap_or("").to_string();
            VerifyResponse { valid: true, subject: sub, code: "OK".into() }
        }
        Err(_) => VerifyResponse { valid: false, subject: "".into(), code: "ERR_INVALID_TOKEN".into() },
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use jsonwebtoken::{encode, EncodingKey, Header};
    use serde_json::json;
    use std::time::{SystemTime, UNIX_EPOCH};

    fn mint(secret: &[u8], sub: &str, exp: u64) -> String {
        let claims = json!({"sub": sub, "exp": exp, "iat": 0});
        encode(&Header::default(), &claims, &EncodingKey::from_secret(secret)).unwrap()
    }

    fn exp_in(s: u64) -> u64 {
        SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() + s
    }

    #[test]
    fn ok() {
        let r = verify_token(&mint(b"test", "alice", exp_in(3600)), b"test");
        assert!(r.valid);
        assert_eq!(r.code, "OK");
        assert_eq!(r.subject, "alice");
    }

    #[test]
    fn empty() {
        let r = verify_token("", b"test");
        assert_eq!(r.code, "ERR_BAD_REQUEST");
    }

    #[test]
    fn bad() {
        let r = verify_token("garbage", b"test");
        assert_eq!(r.code, "ERR_INVALID_TOKEN");
    }

    #[test]
    fn bad_secret() {
        let r = verify_token(&mint(b"good", "alice", exp_in(3600)), b"bad");
        assert_eq!(r.code, "ERR_INVALID_TOKEN");
    }

    #[test]
    fn expired() {
        let r = verify_token(&mint(b"test", "alice", 1), b"test");
        assert_eq!(r.code, "ERR_INVALID_TOKEN");
    }
}
