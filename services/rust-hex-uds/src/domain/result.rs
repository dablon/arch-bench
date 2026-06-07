#[derive(Debug, Clone, PartialEq)]
pub enum DomainCode { Ok, BadRequest, InvalidToken }

impl DomainCode {
    pub fn as_str(&self) -> &'static str {
        match self { Self::Ok => "OK", Self::BadRequest => "ERR_BAD_REQUEST", Self::InvalidToken => "ERR_INVALID_TOKEN" }
    }
}

#[derive(Debug, Clone)]
pub struct VerificationResult { pub valid: bool, pub subject: String, pub code: DomainCode }
