#![allow(unused_variables, dead_code)]

use std::{collections::HashMap, sync::Arc, time::Duration};
use clap::Parser;
use nacos_sdk::api::{constants, naming::{NamingChangeEvent, NamingEventListener, NamingServiceBuilder, ServiceInstance}, props::ClientProps};
use tokio::signal;
use tokio::sync::broadcast;
use tokio::task::JoinHandle;
use crate::server::discovery::{create_naming_service, service_up};
use crate::server::grpc::grpc_serve;
use crate::server::http::http_serve;

mod server;
mod cmd;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt::init();
    let args = cmd::CmdArgs::parse();
    let args_arc = Arc::new(args);

    // http server
    let (shutdown_tx, _) = broadcast::channel(1);
    let hrx = shutdown_tx.subscribe();
    let h_args  = args_arc.clone();
    let http_server_future: JoinHandle<anyhow::Result<()>> = tokio::spawn(async move {
        if let Err(e) = http_serve(&h_args.worker_addr, hrx).await {
            tracing::error!("http server startup error: {}", e);
            return Err(e);
        }
        Ok(())
    });

    // gRPC server
    let grx = shutdown_tx.subscribe();
    let g_args  = args_arc.clone();
    let grpc_server_future: JoinHandle<anyhow::Result<()>> = tokio::spawn(async move {
       if let Err(e) = grpc_serve(g_args.grpc_host, grx).await {
           tracing::error!("grpc server startup error: {}", e);
           return Err(e)
       }
        Ok(())
    });

    // service discovery
    let arg = args_arc.clone();
    let naming_service = create_naming_service(&arg.nacos_addr, &arg.namespace)?;
    let (host, port) = arg.worker_addr.split_once(":").unwrap();
    service_up(&naming_service, host, port.parse()?, arg.grpc_host, &arg.service_name).await?;
    tracing::info!("service up");

    // waiting quit signal
    graceful_shutdown(shutdown_tx).await;
    let results = tokio::join!(http_server_future, grpc_server_future);
    let http_res = results.0.unwrap_or_else(|e| Err(anyhow::anyhow!("HTTP Task Panicked: {}", e)));
    let grpc_res = results.1.unwrap_or_else(|e| Err(anyhow::anyhow!("gRPC Task Panicked: {}", e)));
    if http_res.is_err() || grpc_res.is_err() {
        Err(anyhow::anyhow!("Service exited with error. HTTP: {:?}, GRPC: {:?}", http_res.err(), grpc_res.err()))
    } else {
        Ok(())
    }
}

async fn graceful_shutdown(shutdown_tx: broadcast::Sender<()>) {
    let ctrl_c = async {
        signal::ctrl_c().await.expect("failed to install Ctrl+C handler");
    };
    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };
    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();
    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }
    // 我是希望能在这里 开始关闭 HTTP 和 GRPC 服务
    if shutdown_tx.send(()).is_err() {
        tracing::error!("failed to send shutdown signal");
    }
    tracing::info!("Notify HTTP and gRPC shutdown.");
}


#[tokio::main]
async fn main2() -> anyhow::Result<()>{
    // Nacos client props
    let client_props = ClientProps::new()
        .server_addr("127.0.0.1:8848")
        .namespace("public")
        // .app_name("")
        .auth_username("nacos")
        .auth_password("nacos");

    // Nacos naming services
    let naming_service = NamingServiceBuilder::new(client_props)
    .enable_auth_plugin_http()
    .build()?;

    // 注册服务实例监听器
    let listener = Arc::new(SimpleServiceInstanceChangeListener);
    naming_service.subscribe("works-akamai-web".to_string(), Some(constants::DEFAULT_GROUP.to_string()), Vec::default(), listener).await?;

    // 创建服务实例
    let mut matedata = HashMap::new();
    matedata.insert("gRPC_port".to_string(), "9093".to_string());
    let service_instance = ServiceInstance{
        ip: String::from("192.168.10.95"),
        port: 9093,
        weight: 1.0,
        healthy: true,
        metadata: matedata,
        ..Default::default()
    };
    // 注册服务实例
    naming_service.register_instance(String::from("works-akamai-web"), Some(constants::DEFAULT_GROUP.to_string()), service_instance).await?;
    tokio::time::sleep(Duration::from_secs(15)).await;
    println!("Hello, world!");
    Ok(())
}



/// 服务实例监听器
struct SimpleServiceInstanceChangeListener;
impl NamingEventListener for SimpleServiceInstanceChangeListener {
    fn event(&self, event: std::sync::Arc<NamingChangeEvent>) {
        println!("subscriber service change: {:?}", event);
    }
}