package cmd

import (
	"context"
	"flag"
	"github.com/zlx2019/nacos-go-example/server"
	"github.com/zlx2019/nacos-go-example/tool"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	nacosAddr    string
	nacosHost           = "127.0.0.1"
	nacosPort    uint64 = 8848
	serviceAddr  string
	serviceHost         = tool.LocalIP() // 默认服务地址为内网IP
	gRPCPort     uint64 = 9502           // gRPC 服务端口
	httpPort     uint64 = 8082           // HTTP 服务端口
	instanceName string                  // 注册服务名
)

func init() {
	flag.StringVar(&nacosAddr, "n", os.Getenv("NACOS_ADDR"), "discovery addr, default read env [NACOS_ADDR]")
	flag.StringVar(&serviceAddr, "s", os.Getenv("WORKER_SERVER_ADDR"), "worker service addr, default read env [WORKER_SERVER_ADDR]")
	flag.StringVar(&instanceName, "i", os.Getenv("WORKER_SERVICE_NAME"), "worker service instance name, default read env [WORKER_SERVICE_NAME]")
	flag.Uint64Var(&gRPCPort, "p", gRPCPort, "worker gRPC service port")
}

func Run() {
	flag.Parse()
	if err := parseArgs(); err != nil {
		log.Fatal(err)
	}
	// startup gRPC server
	go func() {
		if err := server.StartupGrpcServer(serviceHost, gRPCPort); err != nil {
			log.Fatalf("Startup grpc server failed: %v", err)
		}
	}()
	// startup HTTP server
	go func() {
		if err := server.StartupHttpServer(serviceHost, httpPort); err != nil {
			log.Fatalf("Startup http server failed: %v", err)
		}
	}()
	// register service discovery
	if err := server.ServiceDiscoveryUp(nacosHost, nacosPort, instanceName, serviceHost, httpPort, gRPCPort); err != nil {
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

func gracefulShutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cleanSig := make(chan struct{})
	go func() {
		if err := server.ServiceDiscoveryDown(instanceName, serviceHost, httpPort); err != nil {
			log.Printf("logout service instance failed: %v", err)
		}
		server.GracefulStopGRPCServer()
		log.Println("grpc server graceful shutdown success")
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("http server shutdown failed: %v", err)
		}
		log.Println("http server graceful shutdown success")
		close(cleanSig)
	}()
	select {
	case <-cleanSig:
		log.Println("graceful shutdown success")
	case <-ctx.Done():
		// 超时, 强制关闭 gRPC 服务端
		log.Println("Shutdown timeout, force exit")
		server.ForceStopGrpcServer()
	}
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
	if serviceAddr != "" {
		host, p, err := net.SplitHostPort(serviceAddr)
		if err != nil {
			log.Fatal(err)
		}
		serviceHost = host
		port, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		gRPCPort = port
	}
	return nil
}
