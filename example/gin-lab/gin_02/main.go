package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// gin_02：路由分组。同一前缀的路由收进一个 Group，中间件可以挂在组上。
func main() {
	router := gin.Default()

	goodsGroup := router.Group("/good")
	{
		goodsGroup.GET("/list", goodsList)
		goodsGroup.GET("/detail", goodsDetail)
		goodsGroup.POST("/add", goodsAdd)
	}

	if err := router.Run(":9022"); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func goodsList(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "goodsList",
	})
}

func goodsDetail(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"id":   ctx.Query("id"),
		"name": "demo goods",
	})
}

func goodsAdd(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"created": true,
	})
}
