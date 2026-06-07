// Composition root for rust-layered-uds.
use std::env;
use std::os::unix::net::UnixListener;
use std::sync::Arc;

use rust_layered_uds::handler::handle;
use rust_layered_uds::repo::JwtRepository;
use rust_layered_uds::service::{TokenVerifierService, VerifierService};

fn main() -> std::io::Result<()> {
    let secret = env::var("JWT_SECRET").unwrap_or_default();
    if secret.is_empty() { eprintln!("JWT_SECRET required"); std::process::exit(2); }
    let path = env::var("UDS_PATH").unwrap_or_else(|_| "/tmp/rust-layered-uds.sock".into());

    let _ = std::fs::remove_file(&path);
    let listener = UnixListener::bind(&path)?;
    std::fs::set_permissions(&path, std::os::unix::fs::PermissionsExt::from_mode(0o660))?;
    eprintln!("rust-layered-uds listening on {}", path);

    let svc: Arc<dyn TokenVerifierService> = Arc::new(VerifierService::new(JwtRepository::new(secret.as_bytes())));

    for stream in listener.incoming() {
        match stream {
            Ok(stream) => {
                let svc = svc.clone();
                std::thread::spawn(move || { handle(stream, svc); });
            }
            Err(e) => eprintln!("accept: {}", e),
        }
    }
    Ok(())
}
