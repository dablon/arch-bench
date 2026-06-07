use crate::repo::{RepoError, TokenRepository};
use std::result::Result;

#[derive(Debug, Clone, PartialEq)]
pub enum ResultCode { Ok, BadRequest, InvalidToken }

impl ResultCode {
    pub fn as_str(&self) -> &'static str {
        match self { Self::Ok => "OK", Self::BadRequest => "ERR_BAD_REQUEST", Self::InvalidToken => "ERR_INVALID_TOKEN" }
    }
}

#[derive(Debug, Clone)]
pub struct VerifyOutcome { pub valid: bool, pub subject: String, pub code: ResultCode }

pub trait TokenVerifierService: Send + Sync { fn verify(&self, token: &str) -> VerifyOutcome; }

pub struct VerifierService<R: TokenRepository> { repo: R }
impl<R: TokenRepository> VerifierService<R> { pub fn new(repo: R) -> Self { Self { repo } } }

impl<R: TokenRepository> TokenVerifierService for VerifierService<R> {
    fn verify(&self, token: &str) -> VerifyOutcome {
        if token.is_empty() {
            return VerifyOutcome { valid: false, subject: String::new(), code: ResultCode::BadRequest };
        }
        match self.repo.verify(token) {
            Ok(sub) => VerifyOutcome { valid: true, subject: sub, code: ResultCode::Ok },
            Err(RepoError::Invalid) => VerifyOutcome { valid: false, subject: String::new(), code: ResultCode::InvalidToken },
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[derive(Clone)]
    struct Fake(Result<String, RepoError>);
    impl TokenRepository for Fake {
        fn verify(&self, _: &str) -> Result<String, RepoError> { self.0.clone() }
    }
    #[test] fn empty() { let s = VerifierService::new(Fake(Ok("x".into()))); assert_eq!(s.verify("").code, ResultCode::BadRequest); }
    #[test] fn ok() { let s = VerifierService::new(Fake(Ok("a".into()))); let r = s.verify("t"); assert!(r.valid); }
    #[test] fn invalid() { let s = VerifierService::new(Fake(Err(RepoError::Invalid))); assert_eq!(s.verify("t").code, ResultCode::InvalidToken); }
}
