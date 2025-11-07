package server

import (
	"context"
	"fmt"
	worker_grpc "github.com/zlx2019/nacos-go-example/pb"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"time"
)

var (
	r          = rand.New(rand.NewSource(time.Now().UnixNano()))
	grpcServer *grpc.Server
)

// WorkerServer gRPC service implement
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

// StartupGrpcServer startup gRPC server
func StartupGrpcServer(host string, port uint64) error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
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

func GracefulStopGRPCServer() {
	if grpcServer != nil {
		grpcServer.GracefulStop()
	}
}

func ForceStopGrpcServer() {
	if grpcServer != nil {
		grpcServer.Stop()
	}
}
