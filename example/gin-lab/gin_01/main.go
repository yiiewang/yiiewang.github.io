package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// gin_01：最小路由。gin.Default() 自带 Logger + Recovery 两个中间件。
func main() {
	router := gin.Default()

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "hello gin",
		})
	})

	if err := router.Run(":9021"); err != nil {
		log.Fatalf("run: %v", err)
	}
}
