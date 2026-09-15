package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// gin_04：query 参数与表单参数。DefaultQuery 给缺省值，PostForm 取表单字段。
func main() {
	router := gin.Default()

	router.GET("/welcome", welcome)
	router.POST("/form", form)

	if err := router.Run(":9024"); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func welcome(ctx *gin.Context) {
	s := ctx.DefaultQuery("firstName", "wel")
	s2 := ctx.DefaultQuery("lastName", "come")
	ctx.JSON(http.StatusOK, gin.H{
		"firstName": s,
		"lastName":  s2,
	})
}

func form(ctx *gin.Context) {
	form := ctx.PostForm("form")
	ctx.JSON(http.StatusOK, gin.H{
		"form": form,
	})
}
