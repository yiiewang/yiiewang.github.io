package server_stub

import (
	"net/rpc"

	"github.com/yiiewang/yiiewang.github.io/example/20260915/rpc-series/05-stub/handler"
)

// HelloServicer 是服务端契约：任何实现了 Hello 方法的类型都能被注册。
type HelloServicer interface {
	Hello(request string, reply *string) error
}

// RegisterHelloService 把“注册”这个动作封进 stub：
// 服务端不再关心注册时用的名字是什么。
func RegisterHelloService(service HelloServicer) error {
	return rpc.RegisterName(handler.HelloServiceName, service)
}
