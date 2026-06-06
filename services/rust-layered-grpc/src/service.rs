// service: business logic. Maps repo errors to domain result codes.
use crate::repo::{RepoError, TokenRepository};

#[derive(Debug, Clone, PartialEq)]
pub enum ResultCode {
    Ok,
    BadRequest,
    InvalidToken,
}

impl ResultCode {
    pub fn as_str(&self) -> &'static str {
        match self {
            ResultCode::Ok => "OK",
            ResultCode::BadRequest => "ERR_BAD_REQUEST",
            ResultCode::InvalidToken => "ERR_INVALID_TOKEN",
        }
    }
}

#[derive(Debug, Clone)]
pub struct VerifyOutcome {
    pub valid: bool,
    pub subject: String,
    pub code: ResultCode,
}

pub trait TokenVerifierService: Send + Sync {
    fn verify(&self, token: &str) -> VerifyOutcome;
}

pub struct VerifierService<R: TokenRepository> {
    repo: R,
}

impl<R: TokenRepository> VerifierService<R> {
    pub fn new(repo: R) -> Self { Self { repo } }
}

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
    use crate::repo::RepoError;

    struct FakeRepo(std::result::Result<String, RepoError>);
    impl TokenRepository for FakeRepo {
        fn verify(&self, _: &str) -> std::result::Result<String, RepoError> {
            match &self.0 {
                Ok(s) => Ok(s.clone()),
                Err(RepoError::Invalid) => Err(RepoError::Invalid),
            }
        }
    }

    #[test]
    fn empty() {
        let s = VerifierService::new(FakeRepo(Ok("x".into())));
        let r = s.verify("");
        assert_eq!(r.code, ResultCode::BadRequest);
    }

    #[test]
    fn ok() {
        let s = VerifierService::new(FakeRepo(Ok("alice".into())));
        let r = s.verify("x");
        assert!(r.valid);
    }

    #[test]
    fn invalid() {
        let s = VerifierService::new(FakeRepo(Err(RepoError::Invalid)));
        let r = s.verify("x");
        assert_eq!(r.code, ResultCode::InvalidToken);
    }
}
