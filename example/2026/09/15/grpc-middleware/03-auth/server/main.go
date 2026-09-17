package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	v1 "github.com/yiiewang/yiiewang.github.io/example/2026/09/15/grpc-middleware/03-auth/proto/gen/interceptor/v1"
)

// 与客户端约定的演示凭据；真实项目从配置中心/环境变量读取
const (
	defaultAppID  = "demo-appid"
	defaultAppKey = "demo-appkey"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type Server struct{}

func (s *Server) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloResponse, error) {
	return &v1.HelloResponse{
		Data: fmt.Sprintf("Hello %s", in.Name),
	}, nil
}

func main() {
	appID := envOr("DEMO_APPID", defaultAppID)
	appKey := envOr("DEMO_APPKEY", defaultAppKey)

	// 在拦截器里做鉴权：拦截器是"每个请求的必经之路"，鉴权放这里最省心
	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		fmt.Printf("接收到一个新的请求: %s\n", info.FullMethod)

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "缺少 metadata")
		}

		// metadata 的 key 会被转成小写
		if got := md.Get("appid"); len(got) == 0 || got[0] != appID {
			return nil, status.Error(codes.Unauthenticated, "APPID 校验失败")
		}
		if got := md.Get("appkey"); len(got) == 0 || got[0] != appKey {
			return nil, status.Error(codes.Unauthenticated, "APPKEY 校验失败")
		}

		i, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		fmt.Println("请求结束")
		return i, nil
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	v1.RegisterHelloServiceServer(s, &Server{})

	l, err := net.Listen("tcp", "127.0.0.1:9013")
	if err != nil {
		log.Fatalf("listen error: %s", err)
	}
	log.Println("auth server listening on 127.0.0.1:9013")

	if err := s.Serve(l); err != nil {
		log.Fatalf("serve error: %s", err)
	}
}
