package main

import (
	"flag"
	"io"
	"log"
	"net"
	"sync"
)

// 最小 TCP 代理：把本地端口的连接转发到目标地址，双向搬运字节。
// 用法：
//
//	go run . -listen=:8080 -target=127.0.0.1:20080
var (
	listen = flag.String("listen", ":8080", "本地监听地址")
	target = flag.String("target", "127.0.0.1:20080", "转发目标地址")
)

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen %s: %v", *listen, err)
	}
	log.Printf("proxy listening on %s -> %s", *listen, *target)

	for {
		src, err := listener.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		log.Printf("accepted from %s", src.RemoteAddr())
		go proxy(src, *target)
	}
}

// proxy 双向搬运两条流：两个方向各一条 goroutine。
// 任何一边读到 EOF，就把“我这半边的写方向”关掉（CloseWrite 发 FIN），
// 让对端知道没有更多数据了 —— 但读方向仍然开着，剩下的数据还能回来。
//
// 常见错误：一边结束就 conn.Close() 整条连接。
// 那会连带掐断另一个方向，长连接场景下表现为“响应还没收完就断开”。
func proxy(src net.Conn, target string) {
	defer src.Close()

	dst, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("dial %s: %v", target, err)
		return
	}
	defer dst.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// 客户端 → 后端
	go func() {
		defer wg.Done()
		if _, err := io.Copy(dst, src); err != nil {
			log.Printf("copy client->backend: %v", err)
		}
		closeWrite(dst)
	}()

	// 后端 → 客户端
	go func() {
		defer wg.Done()
		if _, err := io.Copy(src, dst); err != nil {
			log.Printf("copy backend->client: %v", err)
		}
		closeWrite(src)
	}()

	wg.Wait()
	log.Printf("connection closed: %s", src.RemoteAddr())
}

// closeWrite 只关“写”方向，保留“读”方向。
func closeWrite(conn net.Conn) {
	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.CloseWrite()
	}
}
