package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {

	server := initWebServer()

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "你好，帅气的你来了")
	})

	server.Run(":8077")
}
