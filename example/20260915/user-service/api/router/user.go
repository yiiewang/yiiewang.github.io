package router

import (
	"github.com/gin-gonic/gin"

	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/api/api"
)

func InitUserRouter(router *gin.RouterGroup) {
	userRouter := router.Group("user")
	{
		userRouter.GET("list", api.GetUserList)
		userRouter.POST("register", api.Register)
		userRouter.POST("pwd_login", api.PasswordLogin)
	}
}
