package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// gin_05：参数校验。binding tag 声明规则，ShouldBind 自动按 Content-Type
// 选择 JSON / 表单解析，校验失败时把错误原样返回给调用方。
type SignInForm struct {
	User     string `json:"user" binding:"required,min=3,max=10"`
	Password string `json:"password" binding:"required"`
}

type SignUpForm struct {
	Age      uint8  `form:"age" binding:"gte=1,lte=130"`
	Name     string `form:"name" binding:"required,min=3"`
	Email    string `form:"email" binding:"required,email"`
	Password string `form:"password" binding:"required"`
}

func main() {
	router := gin.Default()

	router.POST("/loginJSON", func(ctx *gin.Context) {
		var loginForm SignInForm
		if err := ctx.ShouldBind(&loginForm); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"msg": "登录成功",
		})
	})

	router.POST("/signUp", func(ctx *gin.Context) {
		var signUpForm SignUpForm
		if err := ctx.ShouldBind(&signUpForm); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"msg": "注册成功",
		})
	})

	if err := router.Run(":9025"); err != nil {
		log.Fatalf("run: %v", err)
	}
}
