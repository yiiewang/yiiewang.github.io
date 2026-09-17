package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// 客户端把“拼 URL + 发请求 + 解 JSON”藏进一个 Add 函数里。
// 调用方只看到 Add(1, 2) —— 这就是 RPC 的雏形：
// 用函数调用的外表，盖住一次网络往返。
type ResponseData struct {
	Data int `json:"data"`
}

func Add(a, b int) int {
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/add?a=%d&b=%d", a, b))
	if err != nil {
		log.Fatalf("http get: %v", err)
	}
	defer resp.Body.Close()

	var data ResponseData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatalf("decode response: %v", err)
	}
	return data.Data
}

func main() {
	fmt.Println(Add(1, 2)) // 3
}
