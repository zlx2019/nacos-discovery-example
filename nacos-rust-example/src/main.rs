use std::{collections::HashMap, sync::Arc, time::Duration};

use nacos_sdk::api::{constants, naming::{NamingChangeEvent, NamingEventListener, NamingServiceBuilder, ServiceInstance}, props::ClientProps};

/// Rust 注册Nacos 

#[tokio::main]
async fn main() -> anyhow::Result<()>{
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