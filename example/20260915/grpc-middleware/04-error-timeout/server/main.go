package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	demo "github.com/yiiewang/yiiewang.github.io/example/20260915/grpc-middleware/04-error-timeout/proto/gen/demo/v1"
)

type HelloService struct{}

// SayHello 演示两条错误路径：
//   - name 为空 → 业务错误，用 status.Error 带上 code
//   - name 为 slow → 慢请求（5s），配合客户端 3s 超时
func (s *HelloService) SayHello(ctx context.Context, in *demo.DemoRequest) (*demo.DemoResponse, error) {
	if in.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name 不能为空")
	}

	if in.Name == "slow" {
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			// 客户端已经放弃，这里没必要再算下去
			return nil, status.Error(codes.Canceled, "客户端已取消")
		}
	}

	return &demo.DemoResponse{
		Message: "hello " + in.Name,
	}, nil
}

func main() {
	s := grpc.NewServer()
	demo.RegisterDemoServiceServer(s, &HelloService{})

	l, err := net.Listen("tcp", "127.0.0.1:9014")
	if err != nil {
		log.Fatalf("listen error: %s", err)
	}
	log.Println("error/timeout server listening on 127.0.0.1:9014")

	if err := s.Serve(l); err != nil {
		log.Fatalf("serve error: %s", err)
	}
}
