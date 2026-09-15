package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	stream "github.com/yiiewang/yiiewang.github.io/example/20260915/grpc-basics/02-stream/proto/v1"
)

type Server struct{}

// GetStream 服务端流：收到一个请求后，持续推送给客户端
func (s *Server) GetStream(req *stream.StreamReqData, res stream.Greeter_GetStreamServer) error {
	if err := res.Send(&stream.StreamResData{
		Data: fmt.Sprintf("req data: %s", req.Data),
	}); err != nil {
		return err
	}

	for {
		if err := res.Send(&stream.StreamResData{
			Data: fmt.Sprintf("get stream: %v", time.Now().Unix()),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
}

// PutStream 客户端流：持续收客户端发来的数据
func (s *Server) PutStream(req stream.Greeter_PutStreamServer) error {
	for {
		srd, err := req.Recv()
		if err != nil {
			return err
		}
		fmt.Println(srd)
	}
}

// AllStream 双向流：一条 goroutine 收、一条 goroutine 发
func (s *Server) AllStream(req stream.Greeter_AllStreamServer) error {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			srd, err := req.Recv()
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
			err := req.Send(&stream.StreamResData{
				Data: fmt.Sprintf("server stream: %v", time.Now().Unix()),
			})
			if err != nil {
				fmt.Printf("send error: %s\n", err)
				return
			}
			time.Sleep(time.Second)
		}
	}()

	wg.Wait()
	return nil
}

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:9002")
	if err != nil {
		log.Fatalf("listen error: %s", err)
	}

	s := grpc.NewServer()
	stream.RegisterGreeterServer(s, &Server{})

	log.Println("stream server listening on 127.0.0.1:9002")
	if err := s.Serve(l); err != nil {
		log.Fatalf("serve error: %s", err)
	}
}
