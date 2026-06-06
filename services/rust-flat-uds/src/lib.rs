pub fn handle_line(line: &str, secret: &[u8]) -> String {
    use jsonwebtoken::{decode, DecodingKey, Validation};
    let line = line.trim_end_matches(|c| c == '\n' || c == '\r');
    if !line.starts_with("VERIFY ") {
        return "ERR BAD_REQUEST\n".to_string();
    }
    let tok = &line[7..];
    if tok.is_empty() {
        return "ERR BAD_REQUEST\n".to_string();
    }
    let mut v = Validation::new(jsonwebtoken::Algorithm::HS256);
    v.leeway = 0;
    match decode::<serde_json::Value>(tok, &DecodingKey::from_secret(secret), &v) {
        Ok(data) => {
            let sub = data.claims.get("sub").and_then(|v| v.as_str()).unwrap_or("");
            format!("OK {}\n", sub)
        }
        Err(_) => "ERR INVALID_TOKEN\n".to_string(),
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
        let tok = mint(b"test", "alice", exp_in(3600));
        let r = handle_line(&format!("VERIFY {}\n", tok), b"test");
        assert_eq!(r, "OK alice\n");
    }

    #[test]
    fn bad_token() {
        let r = handle_line("VERIFY garbage\n", b"test");
        assert_eq!(r, "ERR INVALID_TOKEN\n");
    }

    #[test]
    fn bad_request_no_prefix() {
        let r = handle_line("HELLO\n", b"test");
        assert_eq!(r, "ERR BAD_REQUEST\n");
    }

    #[test]
    fn empty_token() {
        let r = handle_line("VERIFY \n", b"test");
        assert_eq!(r, "ERR BAD_REQUEST\n");
    }

    #[test]
    fn bad_secret() {
        let tok = mint(b"good", "alice", exp_in(3600));
        let r = handle_line(&format!("VERIFY {}\n", tok), b"bad");
        assert_eq!(r, "ERR INVALID_TOKEN\n");
    }

    #[test]
    fn expired() {
        let tok = mint(b"test", "alice", 1);
        let r = handle_line(&format!("VERIFY {}\n", tok), b"test");
        assert_eq!(r, "ERR INVALID_TOKEN\n");
    }

    #[test]
    fn no_newline() {
        let tok = mint(b"test", "alice", exp_in(3600));
        let r = handle_line(&format!("VERIFY {}", tok), b"test");
        assert_eq!(r, "OK alice\n");
    }
}
