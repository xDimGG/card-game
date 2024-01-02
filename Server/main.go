package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	gameServer := NewGameServer(&LobbyManager{})

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("dist/")))
	mux.HandleFunc("/ws", gameServer.handleWs)
	mux.HandleFunc("/chat", handleChatWs)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
