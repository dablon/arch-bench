// rust-hex-uds/src/domain/port.rs: the OUTGOING port.
use crate::domain::result::DomainCode;
use std::result::Result;

pub type RepoResult = Result<String, RepoError>;

#[derive(Debug, Clone)]
pub enum RepoError { Invalid, BadRequest }

impl RepoError { pub fn to_domain(&self) -> DomainCode { match self { Self::Invalid => DomainCode::InvalidToken, Self::BadRequest => DomainCode::BadRequest } } }

pub trait TokenVerifierPort: Send + Sync {
    fn verify(&self, token: &str) -> RepoResult;
}
