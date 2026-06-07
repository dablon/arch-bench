// Library exports for tests. The flat service exposes only the verify
// function (no domain layer) because the test is testing the only
// non-trivial logic.
pub use jsonwebtoken;
use jsonwebtoken::{decode, DecodingKey, Validation};
use serde::{Deserialize, Serialize};

#[derive(Deserialize, Serialize, Debug, PartialEq)]
pub struct VerifyResponse {
    pub valid: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub subject: Option<String>,
    pub code: String,
}

pub fn verify_token(token: &str, secret: &[u8]) -> VerifyResponse {
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

#[cfg(test)]
mod tests {
    use super::*;
    use jsonwebtoken::{encode, EncodingKey, Header};
    use serde_json::json;
    use std::time::{SystemTime, UNIX_EPOCH};

    fn mint(secret: &[u8], sub: &str) -> String {
        let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
        let claims = json!({"sub": sub, "exp": now + 3600});
        encode(&Header::default(), &claims, &EncodingKey::from_secret(secret)).unwrap()
    }

    #[test]
    fn ok() {
        let r = verify_token(&mint(b"test", "alice"), b"test");
        assert!(r.valid);
        assert_eq!(r.code, "OK");
    }

    #[test]
    fn empty() {
        let r = verify_token("", b"test");
        assert_eq!(r.code, "ERR_BAD_REQUEST");
    }

    #[test]
    fn bad() {
        let r = verify_token("x", b"test");
        assert_eq!(r.code, "ERR_INVALID_TOKEN");
    }

    #[test]
    fn bad_secret() {
        let r = verify_token(&mint(b"good", "alice"), b"bad");
        assert_eq!(r.code, "ERR_INVALID_TOKEN");
    }

    #[test]
    fn bad_alg() {
        let r = verify_token("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhbGljZSJ9.", b"test");
        assert_eq!(r.code, "ERR_INVALID_TOKEN");
    }

    #[test]
    fn expired() {
        let claims = json!({"sub": "alice", "exp": 1});
        let tok = encode(&Header::default(), &claims, &EncodingKey::from_secret(b"test")).unwrap();
        let r = verify_token(&tok, b"test");
        assert_eq!(r.code, "ERR_INVALID_TOKEN");
    }
}
