pub mod pb {
    tonic::include_proto!("verifier");
}

// Domain port
pub mod domain {
    pub trait TokenVerifier: Send + Sync {
        fn verify(&self, token: &str) -> Result<String, String>;
    }
}

// OUT adapter
pub mod adapter {
    pub mod jwt {
        use crate::domain::TokenVerifier;
        use jsonwebtoken::{decode, DecodingKey, Validation};

        pub struct JwtVerifier { secret: Vec<u8> }

        impl JwtVerifier {
            pub fn new(secret: &[u8]) -> Self { Self { secret: secret.to_vec() } }
        }

        impl TokenVerifier for JwtVerifier {
            fn verify(&self, token: &str) -> Result<String, String> {
                if token.is_empty() {
                    return Err("ERR_BAD_REQUEST".into());
                }
                let mut v = Validation::new(jsonwebtoken::Algorithm::HS256);
                v.leeway = 0;
                match decode::<serde_json::Value>(token, &DecodingKey::from_secret(&self.secret), &v) {
                    Ok(data) => {
                        let sub = data.claims.get("sub").and_then(|v| v.as_str()).unwrap_or("").to_string();
                        Ok(sub)
                    }
                    Err(_) => Err("ERR_INVALID_TOKEN".into()),
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
                assert_eq!(v.verify(&mint(b"test", "alice", exp_in(3600))).unwrap(), "alice");
            }

            #[test]
            fn empty() {
                assert_eq!(JwtVerifier::new(b"test").verify("").unwrap_err(), "ERR_BAD_REQUEST");
            }

            #[test]
            fn bad() {
                assert_eq!(JwtVerifier::new(b"test").verify("x").unwrap_err(), "ERR_INVALID_TOKEN");
            }

            #[test]
            fn bad_secret() {
                let v = JwtVerifier::new(b"test");
                assert_eq!(v.verify(&mint(b"good", "alice", exp_in(3600))).unwrap_err(), "ERR_INVALID_TOKEN");
            }

            #[test]
            fn expired() {
                assert_eq!(JwtVerifier::new(b"test").verify(&mint(b"test", "alice", 1)).unwrap_err(), "ERR_INVALID_TOKEN");
            }
        }
    }
}

// IN adapter
pub mod transport {
    use super::pb::{verifier_server::Verifier, VerifyRequest, VerifyResponse};
    use super::domain::TokenVerifier;
    use std::sync::Arc;
    use tonic::{Request, Response, Status};

    pub struct GrpcAdapter { v: Arc<dyn TokenVerifier> }

    impl GrpcAdapter {
        pub fn new(v: Arc<dyn TokenVerifier>) -> Self { Self { v } }
    }

    #[tonic::async_trait]
    impl Verifier for GrpcAdapter {
        async fn verify(&self, req: Request<VerifyRequest>) -> Result<Response<VerifyResponse>, Status> {
            let token = req.into_inner().token;
            let resp = match self.v.verify(&token) {
                Ok(sub) => VerifyResponse { valid: true, subject: sub, code: "OK".into() },
                Err(code) => VerifyResponse { valid: false, subject: "".into(), code },
            };
            Ok(Response::new(resp))
        }
    }

    #[cfg(test)]
    mod tests {
        use super::*;
        use crate::adapter::jwt::JwtVerifier;
        use crate::domain::TokenVerifier;
        use jsonwebtoken::{encode, EncodingKey, Header};
        use serde_json::json;
        use std::time::{SystemTime, UNIX_EPOCH};
        use tonic::Request;

        fn exp_in(s: u64) -> u64 {
            SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() + s
        }

        #[tokio::test]
        async fn real_jwt() {
            let v: Arc<dyn TokenVerifier> = Arc::new(JwtVerifier::new(b"test"));
            let g = GrpcAdapter::new(v);
            let claims = json!({"sub": "alice", "exp": exp_in(3600), "iat": 0});
            let tok = encode(&Header::default(), &claims, &EncodingKey::from_secret(b"test")).unwrap();
            let r = g.verify(Request::new(VerifyRequest { token: tok })).await.unwrap().into_inner();
            assert!(r.valid);
            assert_eq!(r.code, "OK");
            assert_eq!(r.subject, "alice");
        }

        #[tokio::test]
        async fn bad() {
            let v: Arc<dyn TokenVerifier> = Arc::new(JwtVerifier::new(b"test"));
            let g = GrpcAdapter::new(v);
            let r = g.verify(Request::new(VerifyRequest { token: "x".into() })).await.unwrap().into_inner();
            assert_eq!(r.code, "ERR_INVALID_TOKEN");
        }

        #[tokio::test]
        async fn empty() {
            let v: Arc<dyn TokenVerifier> = Arc::new(JwtVerifier::new(b"test"));
            let g = GrpcAdapter::new(v);
            let r = g.verify(Request::new(VerifyRequest { token: "".into() })).await.unwrap().into_inner();
            assert_eq!(r.code, "ERR_BAD_REQUEST");
        }
    }
}
