package main

import (
	"crypto/rand"
	"log"
	"net/http"
	"time"
)

var sharedKey = make([]byte, 64)

func main() {
	rand.Read(sharedKey)

	gameServer := NewGameServer(&LobbyManager{})

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("dist/")))
	mux.HandleFunc("/ws", gameServer.handleWs)
	mux.HandleFunc("/chat", handleChatWs)
	mux.HandleFunc("/voice", handleVoiceWs)

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
