// Domain port
pub mod domain {
    #[derive(Debug, PartialEq)]
    pub enum Outcome {
        Ok(String),
        BadRequest,
        Invalid,
    }

    pub trait TokenVerifier: Send + Sync {
        fn verify(&self, token: &str) -> Outcome;
    }
}

// OUT adapter: jwt
pub mod adapter {
    pub mod jwt {
        use crate::domain::{Outcome, TokenVerifier};
        use jsonwebtoken::{decode, DecodingKey, Validation};

        pub struct JwtVerifier { secret: Vec<u8> }

        impl JwtVerifier {
            pub fn new(secret: &[u8]) -> Self { Self { secret: secret.to_vec() } }
        }

        impl TokenVerifier for JwtVerifier {
            fn verify(&self, token: &str) -> Outcome {
                if token.is_empty() {
                    return Outcome::BadRequest;
                }
                let mut v = Validation::new(jsonwebtoken::Algorithm::HS256);
                v.leeway = 0;
                match decode::<serde_json::Value>(token, &DecodingKey::from_secret(&self.secret), &v) {
                    Ok(data) => {
                        let sub = data.claims.get("sub").and_then(|v| v.as_str()).unwrap_or("").to_string();
                        Outcome::Ok(sub)
                    }
                    Err(_) => Outcome::Invalid,
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
                assert_eq!(r, Outcome::Ok("alice".into()));
            }

            #[test]
            fn empty() {
                assert_eq!(JwtVerifier::new(b"test").verify(""), Outcome::BadRequest);
            }

            #[test]
            fn bad() {
                assert_eq!(JwtVerifier::new(b"test").verify("x"), Outcome::Invalid);
            }

            #[test]
            fn bad_secret() {
                let v = JwtVerifier::new(b"test");
                let r = v.verify(&mint(b"good", "alice", exp_in(3600)));
                assert_eq!(r, Outcome::Invalid);
            }

            #[test]
            fn expired() {
                let r = JwtVerifier::new(b"test").verify(&mint(b"test", "alice", 1));
                assert_eq!(r, Outcome::Invalid);
            }
        }
    }
}

// IN adapter: UDS line protocol
pub mod transport {
    use crate::domain::{Outcome, TokenVerifier};

    pub fn handle_line(line: &str, v: &dyn TokenVerifier) -> String {
        let line = line.trim_end_matches(|c| c == '\n' || c == '\r');
        if !line.starts_with("VERIFY ") {
            return "ERR BAD_REQUEST\n".to_string();
        }
        let tok = &line[7..];
        match v.verify(tok) {
            Outcome::Ok(sub) => format!("OK {}\n", sub),
            Outcome::BadRequest => "ERR BAD_REQUEST\n".to_string(),
            Outcome::Invalid => "ERR INVALID_TOKEN\n".to_string(),
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

        fn exp_in(s: u64) -> u64 {
            SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() + s
        }

        #[test]
        fn transport_ok() {
            let v = JwtVerifier::new(b"test");
            let claims = json!({"sub": "alice", "exp": exp_in(3600), "iat": 0});
            let tok = encode(&Header::default(), &claims, &EncodingKey::from_secret(b"test")).unwrap();
            let r = handle_line(&format!("VERIFY {}\n", tok), &v);
            assert_eq!(r, "OK alice\n");
        }

        #[test]
        fn transport_bad_request() {
            let v = JwtVerifier::new(b"test");
            assert_eq!(handle_line("HELLO\n", &v), "ERR BAD_REQUEST\n");
        }

        #[test]
        fn transport_empty_token() {
            let v = JwtVerifier::new(b"test");
            assert_eq!(handle_line("VERIFY \n", &v), "ERR BAD_REQUEST\n");
        }
    }
}
