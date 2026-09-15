package initialize

import (
	my_router "github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/api/router"
	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	router := gin.Default()

	api_group := router.Group("/v1")
	my_router.InitUserRouter(api_group)

	return router
}
