// Domain: Verifier service
pub mod domain {
    use jsonwebtoken::{decode, DecodingKey, Validation};

    pub struct Verifier { secret: Vec<u8> }

    impl Verifier {
        pub fn new(secret: &[u8]) -> Self { Self { secret: secret.to_vec() } }
        pub fn verify(&self, token: &str) -> VerifyOutcome {
            if token.is_empty() {
                return VerifyOutcome::BadRequest;
            }
            let mut v = Validation::new(jsonwebtoken::Algorithm::HS256);
            v.leeway = 0;
            match decode::<serde_json::Value>(token, &DecodingKey::from_secret(&self.secret), &v) {
                Ok(data) => {
                    let sub = data.claims.get("sub").and_then(|v| v.as_str()).unwrap_or("").to_string();
                    VerifyOutcome::Ok(sub)
                }
                Err(_) => VerifyOutcome::Invalid,
            }
        }
    }

    #[derive(Debug, PartialEq)]
    pub enum VerifyOutcome {
        Ok(String),
        BadRequest,
        Invalid,
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
            let v = Verifier::new(b"test");
            assert_eq!(v.verify(&mint(b"test", "alice", exp_in(3600))), VerifyOutcome::Ok("alice".into()));
        }

        #[test]
        fn empty() {
            assert_eq!(Verifier::new(b"test").verify(""), VerifyOutcome::BadRequest);
        }

        #[test]
        fn bad() {
            assert_eq!(Verifier::new(b"test").verify("x"), VerifyOutcome::Invalid);
        }

        #[test]
        fn bad_secret() {
            let v = Verifier::new(b"test");
            assert_eq!(v.verify(&mint(b"good", "alice", exp_in(3600))), VerifyOutcome::Invalid);
        }

        #[test]
        fn expired() {
            let r = Verifier::new(b"test").verify(&mint(b"test", "alice", 1));
            assert_eq!(r, VerifyOutcome::Invalid);
        }
    }
}

// Transport: line-protocol adapter
pub mod transport {
    use crate::domain::{Verifier, VerifyOutcome};

    pub fn handle_line(line: &str, verifier: &Verifier) -> String {
        let line = line.trim_end_matches(|c| c == '\n' || c == '\r');
        if !line.starts_with("VERIFY ") {
            return "ERR BAD_REQUEST\n".to_string();
        }
        let tok = &line[7..];
        match verifier.verify(tok) {
            VerifyOutcome::Ok(sub) => format!("OK {}\n", sub),
            VerifyOutcome::BadRequest => "ERR BAD_REQUEST\n".to_string(),
            VerifyOutcome::Invalid => "ERR INVALID_TOKEN\n".to_string(),
        }
    }

    #[cfg(test)]
    mod tests {
        use super::*;
        use crate::domain::Verifier;
        use jsonwebtoken::{encode, EncodingKey, Header};
        use serde_json::json;
        use std::time::{SystemTime, UNIX_EPOCH};

        fn exp_in(s: u64) -> u64 {
            SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() + s
        }

        #[test]
        fn transport_ok() {
            let v = Verifier::new(b"test");
            let claims = json!({"sub": "alice", "exp": exp_in(3600), "iat": 0});
            let tok = encode(&Header::default(), &claims, &EncodingKey::from_secret(b"test")).unwrap();
            let r = handle_line(&format!("VERIFY {}\n", tok), &v);
            assert_eq!(r, "OK alice\n");
        }

        #[test]
        fn transport_bad_request() {
            let v = Verifier::new(b"test");
            assert_eq!(handle_line("HELLO\n", &v), "ERR BAD_REQUEST\n");
        }

        #[test]
        fn transport_empty_token() {
            let v = Verifier::new(b"test");
            assert_eq!(handle_line("VERIFY \n", &v), "ERR BAD_REQUEST\n");
        }
    }
}
