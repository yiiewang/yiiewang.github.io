package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"

	v1 "github.com/yiiewang/yiiewang.github.io/example/2026/09/15/grpc-middleware/03-auth/proto/gen/interceptor/v1"
)

// 演示用的占位凭据；真实项目从配置中心/环境变量读取
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

// customCredential 实现 grpc.PerRPCCredentials：
// 每次 RPC 前，gRPC 会调用 GetRequestMetadata 拿到键值对，写进请求的 metadata
type customCredential struct{}

func (c customCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"APPID":  envOr("DEMO_APPID", defaultAppID),
		"APPKEY": envOr("DEMO_APPKEY", defaultAppKey),
	}, nil
}

// 本地示例不走 TLS，所以这里返回 false；
// 真实环境应为 true，配合 credentials.TLS 使用
func (c customCredential) RequireTransportSecurity() bool {
	return false
}

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithPerRPCCredentials(customCredential{}))

	cc, err := grpc.Dial("127.0.0.1:9013", opts...)
	if err != nil {
		log.Fatalf("dial error: %s", err)
	}
	defer cc.Close()

	hsc := v1.NewHelloServiceClient(cc)

	reply, err := hsc.SayHello(context.Background(), &v1.HelloRequest{
		Name: "cloaks",
	})
	if err != nil {
		log.Fatalf("call SayHello error: %s", err)
	}

	fmt.Println(reply)
}
