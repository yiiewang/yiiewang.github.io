package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	v1 "github.com/yiiewang/yiiewang.github.io/example/2026/09/15/grpc-middleware/01-metadata/proto/gen/metadata/v1"
)

func main() {
	cc, err := grpc.Dial("127.0.0.1:9011", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial error: %s", err)
	}
	defer cc.Close()

	hsc := v1.NewHelloServerClient(cc)

	// 把键值对挂到 outgoing context 上，gRPC 会把它编码成 HTTP/2 的 header 发出去
	md := metadata.New(map[string]string{
		"name":      "cloaks",
		"timestamp": "good morning",
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	reply, err := hsc.SayHello(ctx, &v1.HelloRequest{
		Name: "cloaks",
	})
	if err != nil {
		log.Fatalf("call SayHello error: %s", err)
	}

	fmt.Println(reply)
}
