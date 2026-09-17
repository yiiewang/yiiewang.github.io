package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	demo "github.com/yiiewang/yiiewang.github.io/example/2026/09/15/grpc-middleware/04-error-timeout/proto/gen/demo/v1"
)

// -name=slow 触发服务端慢请求（5s），客户端 3s 超时；
// -name="" 触发服务端 InvalidArgument。
var name = flag.String("name", "gopher", "请求参数：gopher | slow | 空字符串")

func main() {
	flag.Parse()

	cc, err := grpc.Dial("127.0.0.1:9014", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial error: %s", err)
	}
	defer cc.Close()

	dsc := demo.NewDemoServiceClient(cc)

	// 超时用 context 控制；cancel 一定要调，否则 context 泄漏
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := dsc.SayHello(ctx, &demo.DemoRequest{
		Name: *name,
	})
	if err != nil {
		// gRPC 的错误都带 code：超时是 DeadlineExceeded，业务错误是服务端指定的 code
		s, ok := status.FromError(err)
		if !ok {
			log.Fatalf("非 gRPC 错误: %v", err)
		}
		fmt.Printf("code: %v, message: %v\n", s.Code(), s.Message())
		return
	}

	fmt.Printf("resp.Message: %v\n", resp.Message)
}
