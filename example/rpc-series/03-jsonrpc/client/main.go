package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("dial: %v", err)
	}

	// 客户端一侧同样换成 JSON 编解码器
	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))
	defer client.Close()

	var reply string
	if err := client.Call("HelloService.Hello", "cloaks", &reply); err != nil {
		log.Fatalf("call: %v", err)
	}
	fmt.Println(reply) // hello,cloaks

	// 线上跑的数据长这样（net/rpc/jsonrpc 的请求格式）：
	// {"method":"HelloService.Hello","params":["cloaks"],"id":0}
}
