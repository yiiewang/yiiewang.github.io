package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	v1 "github.com/yiiewang/yiiewang.github.io/example/20260915/grpc-middleware/01-metadata/proto/gen/metadata/v1"
)

type Welcome struct{}

// SayHello 从 incoming metadata 里读客户端随请求带来的键值对。
// metadata 的 key 会被自动转成小写，取值统一用 md.Get(key)。
func (s *Welcome) SayHello(ctx context.Context, request *v1.HelloRequest) (*v1.HelloReply, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "缺少 metadata")
	}

	for key, val := range md {
		fmt.Printf("key: %s, value: %s\n", key, val)
	}

	return &v1.HelloReply{
		Message: fmt.Sprintf("hello %s", request.Name),
	}, nil
}

func main() {
	s := grpc.NewServer()
	v1.RegisterHelloServerServer(s, &Welcome{})

	l, err := net.Listen("tcp", "127.0.0.1:9011")
	if err != nil {
		log.Fatalf("listen error: %s", err)
	}
	log.Println("metadata server listening on 127.0.0.1:9011")

	if err := s.Serve(l); err != nil {
		log.Fatalf("serve error: %s", err)
	}
}
