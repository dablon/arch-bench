// rust-hex-uds/src/infrastructure/jwt_adapter.rs: the OUTGOING adapter.
// Implements the domain port using jsonwebtoken.
use crate::domain::port::{RepoError, RepoResult, TokenVerifierPort};
use jsonwebtoken::{decode, DecodingKey, Validation};
use std::sync::Arc;

pub struct HmacJwtAdapter { secret: Arc<Vec<u8>> }

impl HmacJwtAdapter {
    pub fn new(secret: Vec<u8>) -> Self { Self { secret: Arc::new(secret) } }
}

impl TokenVerifierPort for HmacJwtAdapter {
    fn verify(&self, token: &str) -> RepoResult {
        if token.is_empty() { return Err(RepoError::BadRequest); }
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
    #[test] fn ok() { let a = HmacJwtAdapter::new(b"t".to_vec()); assert_eq!(a.verify(&mint(b"t","alice")).unwrap(), "alice"); }
    #[test] fn bad() { let a = HmacJwtAdapter::new(b"t".to_vec()); assert!(matches!(a.verify("x"), Err(RepoError::Invalid))); }
    #[test] fn empty() { let a = HmacJwtAdapter::new(b"t".to_vec()); assert!(matches!(a.verify(""), Err(RepoError::BadRequest))); }
}
