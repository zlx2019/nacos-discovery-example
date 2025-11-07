use axum::Router;
use axum::routing::get;
use tokio::sync::broadcast;

pub async fn http_serve(addr: &str, rx: broadcast::Receiver<()>) -> anyhow::Result<()>
{
    let app = Router::new()
        .route("/users", get(users));
    let listen = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listen, app)
        .with_graceful_shutdown(blocking_waiting_shutdown(rx))
        .await?;
    Ok(())
}

async fn blocking_waiting_shutdown(mut rx: broadcast::Receiver<()>) {
    let _ = rx.recv().await;
    tracing::info!("[HTTP] received signal, shutting down");
}

async fn users() -> &'static str {
    "ZERO9501"
}