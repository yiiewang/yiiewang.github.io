package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// JSON-RPC 的请求/响应结构。字段必须导出，否则 json.Marshal 只会得到 {}。
type request struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
	ID     int    `json:"id"`
}

type response struct {
	ID     int    `json:"id"`
	Result string `json:"result"`
	Error  any    `json:"error"`
}

func main() {
	body, err := json.Marshal(request{
		Method: "HelloService.Hello",
		Params: []any{"cloaks"},
		ID:     0,
	})
	if err != nil {
		log.Fatalf("marshal request: %v", err)
	}

	// 传输层是 HTTP：POST 一个 JSON 请求体，拿一个 JSON 响应
	resp, err := http.Post("http://localhost:8080/jsonrpc", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("http post: %v", err)
	}
	defer resp.Body.Close()

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		log.Fatalf("decode response: %v", err)
	}
	if r.Error != nil {
		log.Fatalf("rpc error: %v", r.Error)
	}
	fmt.Println(r.Result) // hello,cloaks
}
