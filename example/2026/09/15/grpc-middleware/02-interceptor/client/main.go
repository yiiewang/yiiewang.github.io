package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	v1 "github.com/yiiewang/yiiewang.github.io/example/2026/09/15/grpc-middleware/02-interceptor/proto/gen/interceptor/v1"
)

func main() {
	// 一元客户端拦截器：把真正调用包在中间，可以统计耗时、打日志、注入 metadata
	interceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		fmt.Printf("method: %s, 耗时: %s\n", method, time.Since(start))
		return err
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithUnaryInterceptor(interceptor))

	cc, err := grpc.Dial("127.0.0.1:9012", opts...)
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
