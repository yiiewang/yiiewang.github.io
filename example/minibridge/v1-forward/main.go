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
  v1-forward server -addr :9000
  v1-forward client -listen :8000 -server 127.0.0.1:9000 -target localhost:8000

  server 跑在能访问目标服务的那台机器上
  client 跑在你的本机：监听 -listen，把流量经 -server 转发到 -target
`)
}

// runServer 启动隧道服务端：接受隧道连接，按握手里的目标地址拨号并桥接。
func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	addr := fs.String("addr", ":9000", "隧道监听地址")
	_ = fs.Parse(args)

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen tunnel failed, addr=%s, err=%v", *addr, err)
	}
	log.Printf("tunnel server listening, addr=%s", *addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept failed, err=%v", err)
			continue
		}
		go serveTunnel(conn)
	}
}

// runClient 启动转发客户端：本地监听，每条连接经隧道服务端转发到目标地址。
func runClient(args []string) {
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	listen := fs.String("listen", ":8000", "本地监听地址")
	server := fs.String("server", "127.0.0.1:9000", "隧道服务端地址")
	target := fs.String("target", "", "目标地址 host:port，从服务端视角解析（必填）")
	_ = fs.Parse(args)

	if *target == "" {
		log.Fatalf("missing required flag: -target")
	}

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen local failed, addr=%s, err=%v", *listen, err)
	}
	log.Printf("forward listening, listen=%s, server=%s, target=%s", *listen, *server, *target)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept failed, err=%v", err)
			continue
		}
		go forwardConn(conn, *server, *target)
	}
}
