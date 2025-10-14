package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	httpHandler := InitializeHTTPHandler()
	wsHandler := InitializeWSHandler()
	sseHandler := InitializeSSEHandler()

	r := gin.Default()
	httpHandler.RegisterRoutes(r)

	r.GET("/ws", func(c *gin.Context) {
		wsHandler.HandleWS(c.Writer, c.Request)
	})
	r.GET("/sse", func(c *gin.Context) {
		sseHandler.HandleSSE(c.Writer, c.Request)
	})
}
