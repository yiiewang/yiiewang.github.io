package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/yiiewang/yiiewang.github.io/example/2026/09/15/rpc-series/05-stub/handler"
	"github.com/yiiewang/yiiewang.github.io/example/2026/09/15/rpc-series/05-stub/server_stub"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	// 注册动作交给 server_stub，服务实现只管业务
	if err := server_stub.RegisterHelloService(&handler.NewHelloService{}); err != nil {
		log.Fatalf("register: %v", err)
	}
	log.Println("stub server listening on localhost:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
