package main

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

// 与 02 完全相同的服务实现：换掉的只是“怎么把调用编码成字节”。
// net/rpc 默认用 gob 编码，jsonrpc 把编解码器换成 JSON。
type HelloService struct{}

func (s *HelloService) Hello(request string, reply *string) error {
	*reply = "hello," + request
	return nil
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	if err := rpc.RegisterName("HelloService", &HelloService{}); err != nil {
		log.Fatalf("register: %v", err)
	}
	log.Println("jsonrpc server listening on localhost:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		// ServeCodec：用指定的编解码器服务这条连接，传输层仍是裸 TCP
		go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
