package controller

import "net/http"

type SSEHandler interface {
	HandleSSE(http.ResponseWriter, *http.Request)
}

type implSSEHandler struct {
}

func NewSSEHandler() SSEHandler {
	return &implSSEHandler{}
}

func (h *implSSEHandler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}
