package initialize

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/api/global"
)

// SrvConn 建立到 user_srv 的 gRPC 连接。
// 连接只建一次，复用 HTTP/2 多路复用；每个请求都 Dial 是常见的反模式。
func SrvConn() {
	addr := fmt.Sprintf("%s:%s", global.ServerConfig.UserSrvInfo.Host, global.ServerConfig.UserSrvInfo.Port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Panicf("连接 user_srv 失败: %s", err.Error())
	}
	global.UserSrvConn = conn
}
