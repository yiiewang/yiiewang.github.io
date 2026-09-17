package handler

// 服务名与实现分离：stub 依赖这个常量，服务端/客户端都不再硬编码字符串。
const HelloServiceName = "handler/HelloService"

type NewHelloService struct{}

func (s *NewHelloService) Hello(request string, reply *string) error {
	*reply = "hello," + request
	return nil
}
