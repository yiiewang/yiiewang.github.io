package main

import (
	"io"
	"log"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

// 同一套服务、同一种 JSON 编码，这次把传输层从裸 TCP 换成 HTTP。
// 传输层换了，服务实现一行没动 —— 这就是“编解码器 + 传输”分离的价值。
type HelloService struct{}

func (s *HelloService) Hello(request string, reply *string) error {
	*reply = "hello," + request
	return nil
}

func main() {
	if err := rpc.RegisterName("HelloService", &HelloService{}); err != nil {
		log.Fatalf("register: %v", err)
	}

	http.HandleFunc("/jsonrpc", func(w http.ResponseWriter, r *http.Request) {
		// rpc.ServeRequest 需要的是一个“可读可写可关”的对象：
		// 请求体当输入，ResponseWriter 当输出，拼一个 ReadWriteCloser 给它
		var connect io.ReadWriteCloser = struct {
			io.ReadCloser
			io.Writer
		}{
			ReadCloser: r.Body,
			Writer:     w,
		}
		// 读一个请求、写一个响应，处理完本次调用即返回
		rpc.ServeRequest(jsonrpc.NewServerCodec(connect))
	})

	log.Println("http+jsonrpc server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
