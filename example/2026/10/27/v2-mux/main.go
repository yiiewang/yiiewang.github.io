// Package main 是 minibridge 的第二阶段：一条隧道连接上多路复用任意条流。
//
// 相比 v1 的变化只有"搬运层"：本地监听、dial 目标、双向桥接全都没变，
// 只是把"一条转发一条 TCP"换成"N 条流共享一条 TCP"。
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "server":
		runServer(os.Args[2:])
	case "client":
		runClient(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `用法：
  v2-mux server -addr 127.0.0.1:9000
  v2-mux client -listen :8080 -server 127.0.0.1:9000 -target localhost:8000

  与 v1 的命令行完全一致，区别只在内部：所有转发复用同一条隧道连接。
`)
}

// runServer 启动隧道服务端：每条到来的流都 dial 目标并桥接。
func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:9000", "隧道监听地址")
	_ = fs.Parse(args)

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen tunnel failed, addr=%s, err=%v", *addr, err)
	}
	log.Printf("mux tunnel server listening, addr=%s", *addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept failed, err=%v", err)
			continue
		}
		go serveSession(newSession(conn, false))
	}
}

// serveSession 处理一条隧道连接上的所有流。
func serveSession(sess *Session) {
	for {
		st, err := sess.Accept()
		if err != nil {
			log.Printf("session closed, err=%v", err)
			return
		}
		go handleStream(st)
	}
}

// handleStream 处理对端开出来的一条流：dial 目标 → SYNACK → 桥接。
func handleStream(st *Stream) {
	upstream, err := net.DialTimeout("tcp", st.Target(), dialTimeout)
	if err != nil {
		log.Printf("dial target failed, id=%d, target=%s, err=%v", st.ID(), st.Target(), err)
		_ = st.Reset()
		return
	}

	if err := st.AcceptOK(); err != nil {
		_ = upstream.Close()
		log.Printf("accept stream failed, id=%d, err=%v", st.ID(), err)
		return
	}

	in, out, berr := bridge(st, upstream)
	log.Printf("stream finished, id=%d, target=%s, client_to_target=%d, target_to_client=%d, err=%v",
		st.ID(), st.Target(), in, out, berr)
}

// runClient 启动转发客户端：本地监听，每条连接开一条流。
//
// 与 v1 最大的不同：隧道连接**只建一条**，
// 所有本地连接都在它上面开流，连接数从 1:1 变成 N:1。
func runClient(args []string) {
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	listen := fs.String("listen", ":8080", "本地监听地址")
	server := fs.String("server", "127.0.0.1:9000", "隧道服务端地址")
	target := fs.String("target", "", "目标地址 host:port，从服务端视角解析（必填）")
	_ = fs.Parse(args)

	if *target == "" {
		log.Fatalf("missing required flag: -target")
	}

	remote, err := net.DialTimeout("tcp", *server, dialTimeout)
	if err != nil {
		log.Fatalf("dial tunnel server failed, server=%s, err=%v", *server, err)
	}
	sess := newSession(remote, true)
	log.Printf("mux forward listening, listen=%s, server=%s, target=%s", *listen, *server, *target)

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen local failed, addr=%s, err=%v", *listen, err)
	}

	for {
		local, err := ln.Accept()
		if err != nil {
			log.Printf("accept failed, err=%v", err)
			continue
		}
		go func(local net.Conn) {
			st, err := sess.Open(*target)
			if err != nil {
				log.Printf("open stream failed, target=%s, err=%v", *target, err)
				_ = local.Close()
				return
			}
			in, out, berr := bridge(local, st)
			log.Printf("forward finished, id=%d, target=%s, local_to_remote=%d, remote_to_local=%d, err=%v",
				st.ID(), *target, in, out, berr)
		}(local)
	}
}
