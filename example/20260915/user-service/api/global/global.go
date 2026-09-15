package global

import (
	ut "github.com/go-playground/universal-translator"
	"google.golang.org/grpc"

	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/api/config"
)

var (
	Trans        ut.Translator
	ServerConfig *config.ServerConfig = &config.ServerConfig{}
	UserSrvConn  *grpc.ClientConn
)
