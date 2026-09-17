package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/yiiewang/yiiewang.github.io/example/2026/09/15/grpc-basics/01-unary/proto"
)

type Server struct{}

// SayHello 是一元调用的服务端实现：一个请求对上一个响应
func (s *Server) SayHello(ctx context.Context, request *proto.HelloRequest) (*proto.HelloReply, error) {
	return &proto.HelloReply{
		Message: "Hello " + request.Name,
	}, nil
}

func main() {
	// 创建 server 并注册服务实现
	server := grpc.NewServer()
	proto.RegisterGreeterServer(server, &Server{})

	l, err := net.Listen("tcp", "127.0.0.1:9001")
	if err != nil {
		log.Fatalf("listen error: %s", err)
	}
	log.Println("unary server listening on 127.0.0.1:9001")

	// 阻塞式启动
	if err := server.Serve(l); err != nil {
		log.Fatalf("serve error: %s", err)
	}
}
