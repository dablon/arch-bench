// Domain: ports
pub mod domain {
    use serde::{Deserialize, Serialize};

    #[derive(Deserialize, Serialize, Debug, PartialEq, Clone)]
    pub struct VerifyResponse {
        pub valid: bool,
        #[serde(skip_serializing_if = "Option::is_none")]
        pub subject: Option<String>,
        pub code: String,
    }

    pub trait TokenVerifier: Send + Sync {
        fn verify(&self, token: &str) -> VerifyResponse;
    }
}

// OUT adapter: jwt
pub mod adapter {
    pub mod jwt {
        use crate::domain::{TokenVerifier, VerifyResponse};
        use jsonwebtoken::{decode, DecodingKey, Validation};

        pub struct JwtVerifier { secret: Vec<u8> }

        impl JwtVerifier {
            pub fn new(secret: &[u8]) -> Self { Self { secret: secret.to_vec() } }
        }

        impl TokenVerifier for JwtVerifier {
            fn verify(&self, token: &str) -> VerifyResponse {
                if token.is_empty() {
                    return VerifyResponse { valid: false, subject: None, code: "ERR_BAD_REQUEST".into() };
                }
                let mut v = Validation::new(jsonwebtoken::Algorithm::HS256);
                v.leeway = 0;
                match decode::<serde_json::Value>(token, &DecodingKey::from_secret(&self.secret), &v) {
                    Ok(data) => {
                        let sub = data.claims.get("sub").and_then(|v| v.as_str()).unwrap_or("").to_string();
                        VerifyResponse { valid: true, subject: Some(sub), code: "OK".into() }
                    }
                    Err(_) => VerifyResponse { valid: false, subject: None, code: "ERR_INVALID_TOKEN".into() },
                }
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
                let v = JwtVerifier::new(b"test");
                let r = v.verify(&mint(b"test", "alice", exp_in(3600)));
                assert!(r.valid);
            }

            #[test]
            fn empty() {
                let r = JwtVerifier::new(b"test").verify("");
                assert_eq!(r.code, "ERR_BAD_REQUEST");
            }

            #[test]
            fn bad() {
                let r = JwtVerifier::new(b"test").verify("x");
                assert_eq!(r.code, "ERR_INVALID_TOKEN");
            }

            #[test]
            fn bad_secret() {
                let v = JwtVerifier::new(b"test");
                let r = v.verify(&mint(b"good", "alice", exp_in(3600)));
                assert_eq!(r.code, "ERR_INVALID_TOKEN");
            }

            #[test]
            fn expired() {
                let r = JwtVerifier::new(b"test").verify(&mint(b"test", "alice", 1));
                assert_eq!(r.code, "ERR_INVALID_TOKEN");
            }
        }
    }
}
