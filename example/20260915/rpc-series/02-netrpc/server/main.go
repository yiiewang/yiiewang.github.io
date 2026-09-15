package main

import (
	"log"
	"net"
	"net/rpc"
)

// HelloService 的方法签名要满足 net/rpc 的约定：
// 方法可导出、两个参数（第二个必须是指针）、返回 error。
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

	// 注册后，方法名是 "HelloService.Hello"（服务名 + 方法名）
	if err := rpc.RegisterName("HelloService", &HelloService{}); err != nil {
		log.Fatalf("register: %v", err)
	}
	log.Println("netrpc server listening on localhost:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		// 每条连接一个 goroutine：ServeConn 会阻塞到连接关闭，
		// 不加 go 的话，第二个客户端要等第一个断开才能连上
		go rpc.ServeConn(conn)
	}
}
