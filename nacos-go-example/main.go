package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	worker_grpc "github.com/zlx2019/nacos-go-example/pb"
	"google.golang.org/grpc"
)

var (
	r           = rand.New(rand.NewSource(time.Now().UnixNano()))
	grpcServer  *grpc.Server
	nacosClient naming_client.INamingClient

	nacosAddr string
	nacosHost        = "127.0.0.1"
	nacosPort uint64 = 8848

	workerAddr   string
	workerHost          = LocalIP()
	workerPort   uint64 = 9502
	instanceName string
)

// discovery
type WorkerServer struct {
	worker_grpc.UnimplementedWorkerServiceServer
}

func (s *WorkerServer) Work(ctx context.Context, req *worker_grpc.TaskRequest) (*worker_grpc.TaskReply, error) {
	reply := &worker_grpc.TaskReply{}
	reply.TaskId = req.TaskId
	if r.Intn(100)%2 == 0 {
		reply.Success = true
		reply.Solution = req.Task
		log.Printf("process success: %s \n", req.TaskId)
	} else {
		reply.Success = false
		reply.ErrorCode = "ERROR_NO_SLOT_AVAILABLE"
		reply.ErrorMsg = "服务器资源不足,请稍后重试"
		log.Printf("process failed: %s \n", req.TaskId)
	}
	return reply, nil
}

func init() {
	flag.StringVar(&nacosAddr, "n", os.Getenv("NACOS_ADDR"), "discovery addr, default read env [NACOS_ADDR]")
	flag.StringVar(&workerAddr, "w", os.Getenv("WORKER_ADDR"), "work service addr, default read env [WORKER_ADDR]")
	flag.StringVar(&instanceName, "s", os.Getenv("WORKER_SERVICE_NAME"), "service instance name, default read env [WORKER_SERVICE_NAME]")
}

func main() {
	flag.Parse()
	if err := parseArgs(); err != nil {
		log.Fatal(err)
	}
	// startup gRPC server
	go func() {
		if err := startupGrpcServer(); err != nil {
			log.Fatalf("Startup grpc server failed: %v", err)
		}
	}()
	time.Sleep(time.Second)

	// register service discovery
	if err := registerServiceInstance(); err != nil {
		log.Fatalf("Register service instance failed: %v", err)
	}
	log.Println("Register service instance success.")
	defer gracefulShutdown()

	// wait quit
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	sig := <-quit
	log.Printf("Receive signal: %v, shutdown \n", sig)

}

func parseArgs() error {
	if nacosAddr != "" {
		host, p, err := net.SplitHostPort(nacosAddr)
		if err != nil {
			return err
		}
		nacosHost = host
		port, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return err
		}
		nacosPort = port
	}
	if workerAddr != "" {
		host, p, err := net.SplitHostPort(workerAddr)
		if err != nil {
			log.Fatal(err)
		}
		workerHost = host
		port, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		workerPort = port
	}
	return nil
}

func startupGrpcServer() error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", workerHost, workerPort))
	if err != nil {
		return err
	}
	grpcServer = grpc.NewServer()
	worker_grpc.RegisterWorkerServiceServer(grpcServer, &WorkerServer{})
	if err = grpcServer.Serve(listen); err != nil {
		return err
	}
	return nil
}

func registerServiceInstance() error {
	serverConfig := []constant.ServerConfig{
		*constant.NewServerConfig(nacosHost, nacosPort),
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
		Ip:          workerHost,
		Port:        workerPort,
		ServiceName: instanceName,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"gRPC_port": strconv.Itoa(int(workerPort))},
	})
	if !ok || err != nil {
		return err
	}
	nacosClient = client
	return nil
}

func logoutServiceInstance() error {
	ok, err := nacosClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          workerHost,
		Port:        workerPort,
		ServiceName: instanceName,
		Ephemeral:   true,
	})
	if !ok || err != nil {
		return err
	}
	return nil
}

func gracefulShutdown() {
	// 创建超时上下文, 超过此上下文则强制终止
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cleanSig := make(chan struct{})
	go func() {
		// 注销服务实例
		if err := logoutServiceInstance(); err != nil {
			log.Printf("logout service instance failed: %v", err)
		}
		//time.Sleep(time.Second * 3)
		// 关闭 gRPC 服务端
		if grpcServer != nil {
			grpcServer.GracefulStop()
		}
		close(cleanSig)
	}()
	select {
	case <-cleanSig:
		log.Println("graceful shutdown success")
	case <-ctx.Done():
		// 超时, 强制关闭 gRPC 服务端
		log.Println("Shutdown timeout, force exit")
		if grpcServer != nil {
			grpcServer.Stop()
		}
	}
}
