package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"

	stream "github.com/yiiewang/yiiewang.github.io/example/20260915/grpc-basics/02-stream/proto/v1"
)

// 三种流模式分别演示：-mode=get（服务端流）| put（客户端流）| all（双向流）
var mode = flag.String("mode", "all", "流模式：get | put | all")

func main() {
	flag.Parse()

	cc, err := grpc.Dial("127.0.0.1:9002", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial error: %s", err)
	}
	defer cc.Close()

	gc := stream.NewGreeterClient(cc)

	switch *mode {
	case "get":
		getStream(gc)
	case "put":
		putStream(gc)
	default:
		allStream(gc)
	}
}

// getStream 服务端流：客户端一次请求，服务端持续推送
func getStream(gc stream.GreeterClient) {
	res, err := gc.GetStream(context.Background(), &stream.StreamReqData{
		Data: "cloaks",
	})
	if err != nil {
		log.Fatalf("get stream error: %s", err)
	}
	for {
		srd, err := res.Recv()
		if err != nil {
			log.Printf("recv error: %s", err)
			return
		}
		fmt.Println(srd)
	}
}

// putStream 客户端流：客户端持续发送，服务端只回一次
func putStream(gc stream.GreeterClient) {
	g, err := gc.PutStream(context.Background())
	if err != nil {
		log.Fatalf("put stream error: %s", err)
	}
	for {
		if err := g.Send(&stream.StreamReqData{
			Data: fmt.Sprintf("put stream: %v", time.Now().Unix()),
		}); err != nil {
			log.Printf("send error: %s", err)
			return
		}
		time.Sleep(time.Second)
	}
}

// allStream 双向流：收和发各一条 goroutine
func allStream(gc stream.GreeterClient) {
	g, err := gc.AllStream(context.Background())
	if err != nil {
		log.Fatalf("all stream error: %s", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			srd, err := g.Recv()
			if err != nil {
				fmt.Printf("recv error: %s\n", err)
				return
			}
			fmt.Println(srd)
		}
	}()

	go func() {
		defer wg.Done()
		for {
			err := g.Send(&stream.StreamReqData{
				Data: fmt.Sprintf("client stream: %v", time.Now().Unix()),
			})
			if err != nil {
				fmt.Printf("send error: %s\n", err)
				return
			}
			time.Sleep(time.Second)
		}
	}()

	wg.Wait()
}
