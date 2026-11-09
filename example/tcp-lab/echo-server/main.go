package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
)

// TCP 回声服务器的三种写法，同一个协议（读多少回多少），三种处理“字节流”的姿态。
// 用法：
//
//	go run . -mode=raw     # 显式 read/write：直面短读
//	go run . -mode=iocopy  # io.Copy(conn, conn)：零拷贝环回
//	go run . -mode=bufio   # bufio 按行读：把字节流切成消息
var (
	mode = flag.String("mode", "bufio", "回声模式：raw | iocopy | bufio")
	addr = flag.String("addr", ":20080", "监听地址")
)

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}
	log.Printf("echo server listening on %s (mode=%s)", *addr, *mode)

	var connID atomic.Int64
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		id := connID.Add(1)
		log.Printf("conn-%d: accepted from %s", id, conn.RemoteAddr())

		go func() {
			switch *mode {
			case "raw":
				echoRaw(conn, id)
			case "iocopy":
				echoIOCopy(conn)
			default:
				echoBufio(conn, id)
			}
		}()
	}
}

// echoRaw 最原始的一读一写。
// conn.Read 返回的是“此刻内核缓冲区里有多少读多少”：
// 既不是一条完整消息，也不保证一次读完发送方的全部数据。
func echoRaw(conn net.Conn, id int64) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)

		// 关键顺序：先处理“读到了数据”，再处理错误。
		// Read 允许同时返回 n > 0 和 err == io.EOF（数据已到，连接也关了）。
		if n > 0 {
			msg := fmt.Sprintf("[conn-%d] %s", id, buf[:n])
			if _, werr := conn.Write([]byte(msg)); werr != nil {
				log.Printf("conn-%d: write: %v", id, werr)
				return
			}
		}

		if err == io.EOF {
			log.Printf("conn-%d: client closed", id)
			return
		}
		if err != nil {
			log.Printf("conn-%d: read: %v", id, err)
			return
		}
	}
}

// echoIOCopy 让 io.Copy 自己转。
// conn 同时是源和目标，*net.TCPConn 上 io.Copy 会走 splice(2)：
// 数据全程不进用户态，一行代码换一条零拷贝的回声路径。
func echoIOCopy(conn net.Conn) {
	defer conn.Close()

	if _, err := io.Copy(conn, conn); err != nil {
		log.Printf("iocopy %s: %v", conn.RemoteAddr(), err)
	}
}

// echoBufio 按行回声：bufio.Reader 负责把字节流切成“行”，
// bufio.Writer 负责攒够再发。注意 Flush —— 不 Flush，数据就躺在用户态缓冲区里。
func echoBufio(conn net.Conn, id int64) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		line, err := reader.ReadString('\n')

		if line != "" {
			if _, werr := writer.WriteString(fmt.Sprintf("[conn-%d] %s", id, line)); werr != nil {
				log.Printf("conn-%d: write: %v", id, werr)
				return
			}
			if ferr := writer.Flush(); ferr != nil {
				log.Printf("conn-%d: flush: %v", id, ferr)
				return
			}
		}

		if err != nil {
			if err != io.EOF {
				log.Printf("conn-%d: read: %v", id, err)
			} else {
				log.Printf("conn-%d: client closed", id)
			}
			return
		}
	}
}
