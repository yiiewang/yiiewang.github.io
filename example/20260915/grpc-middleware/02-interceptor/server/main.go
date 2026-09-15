package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	v1 "github.com/yiiewang/yiiewang.github.io/example/20260915/grpc-middleware/02-interceptor/proto/gen/interceptor/v1"
)

type Server struct{}

func (s *Server) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloResponse, error) {
	return &v1.HelloResponse{
		Data: fmt.Sprintf("Hello %s", in.Name),
	}, nil
}

func main() {
	// 一元服务端拦截器：handler 前后各插一段逻辑。
	// 注意 handler 返回的 err 要原样返回——在拦截器里 panic 会拖垮整个进程。
	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		fmt.Printf("接收到一个新的请求: %s\n", info.FullMethod)

		i, err := handler(ctx, req)
		if err != nil {
			fmt.Printf("handler error: %v\n", err)
			return nil, err
		}

		fmt.Println("请求结束")
		return i, nil
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	v1.RegisterHelloServiceServer(s, &Server{})

	l, err := net.Listen("tcp", "127.0.0.1:9012")
	if err != nil {
		log.Fatalf("listen error: %s", err)
	}
	log.Println("interceptor server listening on 127.0.0.1:9012")

	if err := s.Serve(l); err != nil {
		log.Fatalf("serve error: %s", err)
	}
}
