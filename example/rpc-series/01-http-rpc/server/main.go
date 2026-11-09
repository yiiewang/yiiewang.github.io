package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// 最小的“远程调用”：HTTP 当传输协议，query 传参，JSON 回包。
// 调用方看到的只是一个 Add 函数，函数体里却发生了一次跨进程请求。
//
// 协议约定（这是本节全部的设计）：
//   - callID：URL path（/add）
//   - 参数：query string（a=1&b=2）
//   - 编码：JSON
func main() {
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		a, errA := strconv.Atoi(r.Form.Get("a"))
		b, errB := strconv.Atoi(r.Form.Get("b"))
		if errA != nil || errB != nil {
			http.Error(w, "a/b 必须是整数", http.StatusBadRequest)
			return
		}

		fmt.Println("path:", r.URL.Path) // 服务端视角：每次调用就是一次普通 HTTP 请求

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"data": a + b})
	})

	log.Println("http-rpc server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
