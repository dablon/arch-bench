use clap::Parser;
use jsonwebtoken::{encode, EncodingKey, Header};
use serde::{Deserialize, Serialize};
use std::time::{SystemTime, UNIX_EPOCH};

#[derive(Debug, Serialize, Deserialize)]
struct Claims {
    sub: String,
    exp: u64,
    iat: u64,
    role: String,
}

#[derive(Parser, Debug)]
#[command(name = "token-gen", about = "Generate a JWT for arch-bench services")]
struct Args {
    /// HMAC secret
    #[arg(long, default_value = "")]
    secret: String,
    /// Subject (user id)
    #[arg(long, default_value = "user-1")]
    sub: String,
    /// Role
    #[arg(long, default_value = "user")]
    role: String,
    /// Expiry in seconds from now
    #[arg(long, default_value_t = 3600)]
    ttl: u64,
    /// Read secret from env JWT_SECRET
    #[arg(long, default_value_t = false)]
    from_env: bool,
}

fn main() {
    let args = Args::parse();
    let secret = if args.from_env {
        std::env::var("JWT_SECRET").unwrap_or_default()
    } else {
        args.secret
    };
    if secret.is_empty() {
        eprintln!("error: empty secret (use --secret or --from-env with JWT_SECRET set)");
        std::process::exit(2);
    }
    let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
    let claims = Claims {
        sub: args.sub,
        exp: now + args.ttl,
        iat: now,
        role: args.role,
    };
    let token = encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(secret.as_bytes()),
    )
    .unwrap();
    println!("{}", token);
}
