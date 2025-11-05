import asyncio
import os

from v2.nacos import ClientConfigBuilder, GRPCConfig, ClientConfig, NacosNamingService, RegisterInstanceParam, \
    DeregisterInstanceParam


async def main():
    config = client_config("127.0.0.1:8848")
    client = await naming_client(config)
    if await register_service(client, "worker-other", "192.168.10.95", 9503):
        print("注册成功")
    await asyncio.sleep(10)
    if await logout_service(client, "worker-other", "192.168.10.95", 9503):
        print("注销成功")
    await asyncio.sleep(5)


## 注册服务实例
async def register_service(client: NacosNamingService, service_name: str, ip: str, port: int) -> bool:
    try:
        return await client.register_instance(request=RegisterInstanceParam(
            service_name=service_name,
            ip=ip,
            port=port,
            metadata={'gRPC_port': '9503'},
        ))
    except Exception as e:
        print(f"注册服务实例错误: {e}")
        return False

## 注销服务
async def logout_service(client: NacosNamingService, service_name: str, ip: str, port: int) -> bool:
    return await client.deregister_instance(request=DeregisterInstanceParam(
        service_name=service_name,
        ip=ip,
        port=port,
    ))


### 构建客户端
async def naming_client(config: ClientConfig) -> NacosNamingService:
    return await NacosNamingService.create_naming_service(config)


### 构建客户端配置
def client_config(server_addr: str) -> ClientConfig:
    return (ClientConfigBuilder()
            .server_address(os.getenv('NACOS_ADDR', server_addr))
            .namespace_id("ez-platform")
            .log_level('INFO')
            .grpc_config(GRPCConfig(grpc_timeout=5000))
            .build())


if __name__ == '__main__':
    asyncio.run(main())
