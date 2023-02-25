package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	gameServer := NewGameServer(&LobbyManager{})

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("dist/")))
	mux.HandleFunc("/ws", gameServer.handleWs)

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
