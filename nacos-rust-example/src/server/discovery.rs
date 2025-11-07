use nacos_sdk::api::constants;
use nacos_sdk::api::naming::{NamingService, NamingServiceBuilder, ServiceInstance};
use nacos_sdk::api::props::ClientProps;
use std::collections::HashMap;


pub fn create_naming_service(addr: &str, namespace: &str) -> anyhow::Result<NamingService>{
    let client_props = ClientProps::new()
        .server_addr(addr)
        .namespace(namespace);
     Ok(NamingServiceBuilder::new(client_props)
         .enable_auth_plugin_http()
         .build()?)
}

pub async fn service_up(naming_service: &NamingService,
              ip: &str, port: i32, grpc_port: i32,
              service_name: &str) -> anyhow::Result<()> {
    let instance = ServiceInstance{
        ip: ip.to_string(),
        port,
        weight: 1.0,
        healthy: true,
        enabled: true,
        ephemeral: true,
        metadata: HashMap::from([("gRPC_port".to_string(), grpc_port.to_string())]),
        ..Default::default()
    };
    naming_service.register_instance(service_name.to_string(), Some(constants::DEFAULT_GROUP.to_string()), instance).await?;
    Ok(())
}

fn service_down(naming_service: &NamingService){
    todo!()
}
