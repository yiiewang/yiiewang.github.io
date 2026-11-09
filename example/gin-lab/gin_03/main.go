package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// gin_03：路径参数绑定。ShouldBindUri 把 /good/:id/:action 里的字段填进结构体，
// 并做 binding 校验，校验失败直接 404。
type Good struct {
	ID     int    `uri:"id" binding:"required"`
	Action string `uri:"action" binding:"required"`
}

func main() {
	router := gin.Default()

	goodsGroup := router.Group("/good")
	{
		goodsGroup.GET("", goodsList)
		goodsGroup.GET("/:id/:action", goodsDetail)
		goodsGroup.POST("", goodsAdd)
	}

	if err := router.Run(":9023"); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func goodsList(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "goodsList",
	})
}

func goodsDetail(ctx *gin.Context) {
	var good Good
	if err := ctx.ShouldBindUri(&good); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":     good.ID,
		"action": good.Action,
	})
}

func goodsAdd(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"created": true,
	})
}
