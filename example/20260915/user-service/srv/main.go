package main

import (
	"flag"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/proto"
	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/srv/global"
	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/srv/handler"
	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/srv/model"
)

func main() {
	// 开发期让表跟上结构体；生产迁移请用 golang-migrate / goose
	if err := global.DB.AutoMigrate(&model.User{}); err != nil {
		panic(err)
	}

	// 获取命令行 flag
	IP := flag.String("ip", "0.0.0.0", "ip 地址")
	Port := flag.Int("port", 8080, "端口号")
	flag.Parse()

	fmt.Printf("IP: %v\n", *IP)
	fmt.Printf("Port: %v\n", *Port)

	s := grpc.NewServer()
	proto.RegisterUserServer(s, &handler.UserServer{})
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic(err)
	}
	err = s.Serve(l)
	if err != nil {
		panic(err)
	}
}
