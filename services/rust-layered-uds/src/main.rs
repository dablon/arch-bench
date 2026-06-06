use std::env;
use std::io::{BufRead, BufReader, Write};
use std::os::unix::net::UnixListener;
use std::sync::Arc;

use rust_layered_uds::domain::Verifier;
use rust_layered_uds::transport::handle_line;

fn main() {
    let secret = env::var("JWT_SECRET").unwrap_or_default();
    if secret.is_empty() {
        eprintln!("JWT_SECRET required");
        std::process::exit(2);
    }
    let path = env::var("UDS_PATH").unwrap_or_else(|_| "/tmp/rust-layered-uds.sock".into());
    let _ = std::fs::remove_file(&path);
    let listener = UnixListener::bind(&path).expect("bind");
    let _ = std::fs::set_permissions(&path, std::os::unix::fs::PermissionsExt::from_mode(0o660));
    let verifier = Arc::new(Verifier::new(secret.as_bytes()));
    eprintln!("rust-layered-uds listening on {}", path);
    for stream in listener.incoming() {
        let mut s = match stream { Ok(s) => s, Err(_) => continue };
        let v = verifier.clone();
        std::thread::spawn(move || {
            let r = BufReader::new(s.try_clone().unwrap());
            for line in r.lines() {
                let line = match line { Ok(l) => l, Err(_) => break };
                let resp = handle_line(&line, &v);
                let _ = s.write_all(resp.as_bytes());
            }
        });
    }
}
