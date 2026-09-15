package client_stub

import (
	"fmt"
	"net/rpc"

	"github.com/yiiewang/yiiewang.github.io/example/20260915/rpc-series/05-stub/handler"
)

// HelloServiceStub 把 rpc.Client 包一层，对外只暴露业务方法。
// 调用方看到的是 client.Hello("cloaks", &reply)，不是 client.Call("xxx.Hello", ...)。
type HelloServiceStub struct {
	*rpc.Client
}

func NewHelloServiceClient(protocol string, address string) HelloServiceStub {
	client, err := rpc.Dial(protocol, address)
	if err != nil {
		panic(fmt.Sprintf("dial error: %s", err))
	}
	return HelloServiceStub{client}
}

func (c *HelloServiceStub) Hello(request string, reply *string) error {
	return c.Call(handler.HelloServiceName+".Hello", request, reply)
}
