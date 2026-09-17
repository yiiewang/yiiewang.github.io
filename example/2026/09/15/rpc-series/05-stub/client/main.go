package main

import (
	"fmt"
	"log"

	"github.com/yiiewang/yiiewang.github.io/example/2026/09/15/rpc-series/05-stub/client_stub"
)

func main() {
	// 调用方拿到的是一个“本地对象”，Hello 的调用过程被 stub 完全盖住
	client := client_stub.NewHelloServiceClient("tcp", "localhost:8080")

	var reply string
	if err := client.Hello("cloaks", &reply); err != nil {
		log.Fatalf("call: %v", err)
	}
	fmt.Println(reply) // hello,cloaks
}
