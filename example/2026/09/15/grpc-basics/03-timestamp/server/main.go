package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	timestamp "github.com/yiiewang/yiiewang.github.io/example/2026/09/15/grpc-basics/03-timestamp/proto/gen/timestamp/v1"
)

type Hello struct{}

func (h *Hello) SayHello(ctx context.Context, in *timestamp.HelloRequest) (*timestamp.HelloResponse, error) {
	return &timestamp.HelloResponse{
		Data: fmt.Sprintf("gender is %d and timestamp is %s", in.Gender, in.CreateTime),
	}, nil
}

func main() {
	s := grpc.NewServer()
	timestamp.RegisterHelloServerServer(s, &Hello{})

	l, err := net.Listen("tcp", "127.0.0.1:9003")
	if err != nil {
		log.Fatalf("listen error: %s", err)
	}
	log.Println("timestamp server listening on 127.0.0.1:9003")

	if err := s.Serve(l); err != nil {
		log.Fatalf("serve error: %s", err)
	}
}
