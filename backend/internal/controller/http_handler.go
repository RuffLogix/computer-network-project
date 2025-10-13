package controller

type HttpHandler interface {
}

type implHttpHandler struct {
}

func NewHttpHander() HttpHandler {
	return &implHttpHandler{}
}
