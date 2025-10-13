package main

import (
	"log"
	"net/http"
)

func main() {
	wsHandler := InitializeWSHandler()

	http.HandleFunc("/ws", wsHandler.HandleWS)
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
