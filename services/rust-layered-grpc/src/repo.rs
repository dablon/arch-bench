// repo: data access. JWT decoding and signature verification. No
// gRPC, no proto types, no business knowledge.
use jsonwebtoken::{decode, DecodingKey, Validation};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum RepoError {
    #[error("invalid token")]
    Invalid,
}

pub trait TokenRepository: Send + Sync {
    fn verify(&self, token: &str) -> Result<String, RepoError>;
}

pub struct JwtRepository {
    secret: Vec<u8>,
}

impl JwtRepository {
    pub fn new(secret: &[u8]) -> Self { Self { secret: secret.to_vec() } }
}

impl TokenRepository for JwtRepository {
    fn verify(&self, token: &str) -> Result<String, RepoError> {
        let v = Validation::new(jsonwebtoken::Algorithm::HS256);
        match decode::<serde_json::Value>(token, &DecodingKey::from_secret(&self.secret), &v) {
            Ok(data) => {
                let sub = data.claims.get("sub").and_then(|v| v.as_str()).unwrap_or("").to_string();
                Ok(sub)
            }
            Err(_) => Err(RepoError::Invalid),
        }
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
        let r = JwtRepository::new(b"test");
        assert_eq!(r.verify(&mint(b"test", "alice")).unwrap(), "alice");
    }

    #[test]
    fn bad() {
        let r = JwtRepository::new(b"test");
        assert!(matches!(r.verify("x"), Err(RepoError::Invalid)));
    }

    #[test]
    fn bad_alg() {
        let r = JwtRepository::new(b"test");
        assert!(matches!(r.verify("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhbGljZSJ9."), Err(RepoError::Invalid)));
    }

    #[test]
    fn bad_secret() {
        let r = JwtRepository::new(b"test");
        assert!(matches!(r.verify(&mint(b"good", "alice")), Err(RepoError::Invalid)));
    }
}
