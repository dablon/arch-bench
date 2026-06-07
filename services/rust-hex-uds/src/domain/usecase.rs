use crate::domain::port::TokenVerifierPort;
use crate::domain::result::{DomainCode, VerificationResult};

pub fn verify(port: &dyn TokenVerifierPort, token: &str) -> VerificationResult {
    if token.is_empty() {
        return VerificationResult { valid: false, subject: String::new(), code: DomainCode::BadRequest };
    }
    match port.verify(token) {
        Ok(sub) => VerificationResult { valid: true, subject: sub, code: DomainCode::Ok },
        Err(e) => VerificationResult { valid: false, subject: String::new(), code: e.to_domain() },
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::port::{RepoError, RepoResult};

    struct Fake(Result<String, RepoError>);
    impl TokenVerifierPort for Fake { fn verify(&self, _: &str) -> RepoResult { self.0.clone() } }

    #[test] fn empty() { let p = Fake(Ok("x".into())); assert_eq!(verify(&p, "").code, DomainCode::BadRequest); }
    #[test] fn ok() { let p = Fake(Ok("a".into())); let r = verify(&p, "t"); assert!(r.valid); }
    #[test] fn invalid() { let p = Fake(Err(RepoError::Invalid)); assert_eq!(verify(&p, "t").code, DomainCode::InvalidToken); }
}
