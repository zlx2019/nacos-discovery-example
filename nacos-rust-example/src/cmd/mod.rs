use std::net::SocketAddr;

use clap::Parser;


/// Program args
#[derive(Parser, Debug, Clone)]
pub struct CmdArgs {
    /// nacos server addr
    #[arg(short, long, default_value = "127.0.0.1:8848", value_parser = parse_addr)]
    pub nacos_addr: String,
    /// nacos service namespace
    #[arg(long, default_value = "ez-platform")]
    pub namespace: String,
    /// worker service host
    #[arg(short, long, default_value = "127.0.0.1:9505", value_parser = parse_addr)]
    pub worker_addr: String,
    /// gRPC service port
    #[arg(long, default_value = "9514")]
    pub grpc_host: i32,
    /// service discovery name
    #[arg(long, default_value = "worker-other")]
    pub service_name: String,
}

fn parse_addr(addr: &str) -> Result<String, String>{
    match addr.parse::<SocketAddr>() {
        Ok(_) => {
            Ok(String::from(addr))
        },
        Err(e) => {
            Err("Invalid address".to_string())
        }
    }
}