package controller

type SSEHandler interface {
}

type impleSSEHandler struct {
}

func NewSSEHandler() SSEHandler {
	return &impleSSEHandler{}
}
