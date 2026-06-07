use rust_hex_uds::infrastructure::uds_server::{build_port, handle};
use std::env;
use std::os::unix::net::UnixListener;

fn main() -> std::io::Result<()> {
    let secret = env::var("JWT_SECRET").unwrap_or_default();
    if secret.is_empty() { eprintln!("JWT_SECRET required"); std::process::exit(2); }
    let path = env::var("UDS_PATH").unwrap_or_else(|_| "/tmp/rust-hex-uds.sock".into());

    let _ = std::fs::remove_file(&path);
    let listener = UnixListener::bind(&path)?;
    std::fs::set_permissions(&path, std::os::unix::fs::PermissionsExt::from_mode(0o660))?;
    eprintln!("rust-hex-uds listening on {}", path);

    let port = build_port(secret.into_bytes());

    for stream in listener.incoming() {
        match stream {
            Ok(stream) => {
                let port = port.clone();
                std::thread::spawn(move || { handle(stream, port); });
            }
            Err(e) => eprintln!("accept: {}", e),
        }
    }
    Ok(())
}
