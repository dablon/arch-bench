use jsonwebtoken::{decode, DecodingKey, Validation};
use thiserror::Error;
use std::result::Result;

#[derive(Debug, Clone, Error)]
pub enum RepoError { #[error("invalid token")] Invalid }

pub trait TokenRepository: Send + Sync {
    fn verify(&self, token: &str) -> Result<String, RepoError>;
}

pub struct JwtRepository { secret: Vec<u8> }
impl JwtRepository { pub fn new(secret: &[u8]) -> Self { Self { secret: secret.to_vec() } } }

impl TokenRepository for JwtRepository {
    fn verify(&self, token: &str) -> Result<String, RepoError> {
        let v = Validation::new(jsonwebtoken::Algorithm::HS256);
        match decode::<serde_json::Value>(token, &DecodingKey::from_secret(&self.secret), &v) {
            Ok(d) => Ok(d.claims.get("sub").and_then(|v| v.as_str()).unwrap_or("").to_string()),
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
    fn mint(s: &[u8], sub: &str) -> String {
        let n = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
        let c = json!({"sub": sub, "exp": n + 3600});
        encode(&Header::default(), &c, &EncodingKey::from_secret(s)).unwrap()
    }
    #[test] fn ok() { assert_eq!(JwtRepository::new(b"t").verify(&mint(b"t","alice")).unwrap(), "alice"); }
    #[test] fn bad() { assert!(JwtRepository::new(b"t").verify("x").is_err()); }
    #[test] fn bad_secret() { assert!(JwtRepository::new(b"t").verify(&mint(b"x","alice")).is_err()); }
    #[test] fn bad_alg() { assert!(JwtRepository::new(b"t").verify("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhbGljZSJ9.").is_err()); }
}
