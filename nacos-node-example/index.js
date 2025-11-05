import { NacosNamingClient } from "nacos";

async function main() {
  let client = createNamingClient("127.0.0.1:8848", "ez-platform");
  await registerService(client, "worker-other", "192.168.10.95", 9504)
    .then(() => {
      console.log("注册成功");
    })
    .catch((e) => {
      console.log("注册失败");
    });
}

async function registerService(client, serviceName, ip, port) {
  return client.registerInstance(serviceName, {
    ip: ip,
    port: port,
    serviceName: serviceName,
    healthy: true,
    ephemeral: true,
    enabled: true,
  });
}

// create Naming client
function createNamingClient(nacosAddr, namespace) {
  const client = new NacosNamingClient({
    logger: console,
    serverList: nacosAddr,
    namespace: namespace,
  });
  client.ready();
  return client;
}

main();
