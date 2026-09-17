package main

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/yiiewang/yiiewang.github.io/example/2026/09/15/user-service/api/global"
	"github.com/yiiewang/yiiewang.github.io/example/2026/09/15/user-service/api/initialize"
)

func main() {
	// initialize zap logger
	initialize.Zap()
	// initialize config file
	initialize.Config()
	// initialize gRPC connection to user_srv
	initialize.SrvConn()
	// initialize router
	Router := initialize.Routers()
	// initialize translator
	if err := initialize.Trans("zh"); err != nil {
		zap.S().Panicln(err)
	}

	zap.S().Infof("服务器启动，端口：%s", global.ServerConfig.Port)
	if err := Router.Run(fmt.Sprintf(":%s", global.ServerConfig.Port)); err != nil {
		zap.S().Panic("启动失败：", err.Error())
	}
}
