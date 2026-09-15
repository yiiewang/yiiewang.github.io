package main

import (
	"flag"
	"log"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

// 并发端口扫描器：worker 池 + 连接超时。
// 只扫描你自己有权限测试的主机。
//
// 用法：
//
//	go run . -host=127.0.0.1 -start=1 -end=1024 -workers=256
var (
	host    = flag.String("host", "127.0.0.1", "目标主机")
	start   = flag.Int("start", 1, "起始端口")
	end     = flag.Int("end", 1024, "结束端口")
	workers = flag.Int("workers", 256, "并发 worker 数")
	timeout = flag.Duration("timeout", 500*time.Millisecond, "单端口连接超时")
)

func main() {
	flag.Parse()

	if *start < 1 || *end > 65535 || *start > *end {
		log.Fatalf("端口范围非法: %d-%d", *start, *end)
	}

	startAt := time.Now()

	// 任务通道：容量给满，生产者不必等消费者
	ports := make(chan int, *end-*start+1)
	// 结果通道同样带缓冲：worker 投递结果时不会卡住
	open := make(chan int, *end-*start+1)

	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range ports {
				if scan(*host, port, *timeout) {
					open <- port
				}
			}
		}()
	}

	// 生产者：把所有端口喂给 worker 池，然后关闭任务通道（worker 靠它退出）
	for p := *start; p <= *end; p++ {
		ports <- p
	}
	close(ports)

	// 等待所有 worker 结束，再关闭结果通道，主 goroutine 才能安全地 range
	go func() {
		wg.Wait()
		close(open)
	}()

	var openPorts []int
	for p := range open {
		openPorts = append(openPorts, p)
	}
	sort.Ints(openPorts)

	log.Printf("scanned %s:%d-%d in %v, %d open: %v",
		*host, *start, *end, time.Since(startAt).Round(time.Millisecond), len(openPorts), openPorts)
}

// scan 用一次完整的 TCP 三次握手判断端口是否开放：
// DialTimeout 成功即开放，超时/拒绝即关闭（或不可达）。
func scan(host string, port int, timeout time.Duration) bool {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
