package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/yiiewang/yiiewang.github.io/example/2026/09/15/grpc-basics/03-timestamp/proto/gen/timestamp/v1"
)

func main() {
	cc, err := grpc.Dial("127.0.0.1:9003", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial error: %s", err)
	}
	defer cc.Close()

	hsc := v1.NewHelloServerClient(cc)

	// enum 直接当常量用；时间戳用标准库 timestamppb 包装
	res, err := hsc.SayHello(context.Background(), &v1.HelloRequest{
		Gender:     v1.Gender_FEMALE,
		CreateTime: timestamppb.New(time.Now()),
	})
	if err != nil {
		log.Fatalf("call SayHello error: %s", err)
	}

	fmt.Println(res.Data)
}
