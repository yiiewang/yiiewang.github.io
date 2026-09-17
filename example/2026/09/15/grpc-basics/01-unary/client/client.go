package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"

	"github.com/yiiewang/yiiewang.github.io/example/2026/09/15/grpc-basics/01-unary/proto"
)

func main() {
	// 建立连接（本地示例不开 TLS）
	dial, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial error: %s", err)
	}
	defer dial.Close()

	client := proto.NewGreeterClient(dial)

	// 像调用本地方法一样调用远程方法
	reply, err := client.SayHello(context.Background(), &proto.HelloRequest{
		Name: "cloaks",
	})
	if err != nil {
		log.Fatalf("call SayHello error: %s", err)
	}
	fmt.Println(reply)
}
