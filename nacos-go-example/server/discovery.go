package server

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"strconv"
)

var nacosClient naming_client.INamingClient

// ServiceDiscoveryUp 服务上线
func ServiceDiscoveryUp(addr string, port uint64, serviceName string, ip string, httpPort, grpcPort uint64) error {
	serverConfig := []constant.ServerConfig{
		*constant.NewServerConfig(addr, port),
	}
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId("ez-platform"),
		constant.WithTimeoutMs(5000),
		constant.WithLogLevel("debug"),
	)
	client, err := clients.NewNamingClient(vo.NacosClientParam{
		ServerConfigs: serverConfig,
		ClientConfig:  &clientConfig,
	})
	if err != nil {
		return err
	}
	// 注册服务实例
	ok, err := client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        httpPort,
		ServiceName: serviceName,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"gRPC_port": strconv.Itoa(int(grpcPort))},
	})
	if !ok || err != nil {
		return err
	}
	nacosClient = client
	return nil
}

// ServiceDiscoveryDown 服务下线
func ServiceDiscoveryDown(serviceName string, ip string, httpPort uint64) error {
	ok, err := nacosClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        httpPort,
		ServiceName: serviceName,
		Ephemeral:   true,
	})
	if !ok || err != nil {
		return err
	}
	return nil
}
