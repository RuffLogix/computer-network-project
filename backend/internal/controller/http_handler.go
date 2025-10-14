package controller

import "github.com/gin-gonic/gin"

type HTTPHandler interface {
	RegisterRoutes(*gin.Engine)
}

type implHTTPHandler struct {
}

func NewHTTPHander() HTTPHandler {
	return &implHTTPHandler{}
}

func (h *implHTTPHandler) RegisterRoutes(router *gin.Engine) {

}
