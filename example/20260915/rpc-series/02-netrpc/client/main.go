package main

import (
	"fmt"
	"log"
	"net/rpc"
)

func main() {
	client, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer client.Close()

	// 调用格式："服务名.方法名"，参数与返回值由 net/rpc 自动编解码
	var reply string
	if err := client.Call("HelloService.Hello", "cloaks", &reply); err != nil {
		log.Fatalf("call: %v", err)
	}
	fmt.Println(reply) // hello,cloaks
}
