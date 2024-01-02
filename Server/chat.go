package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var chatUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// map of chatLobbies to clients to connections
var chatLobbies = map[string]map[string]*websocket.Conn{}

func handleChatClose(lobbyID string, clientID string) {
	if lobby, ok := chatLobbies[lobbyID]; ok {
		if client, ok := lobby[clientID]; ok {
			client.Close()
			delete(lobby, clientID)
		}

		// Remove empty lobbies from map
		if len(lobby) == 0 {
			delete(chatLobbies, lobbyID)
		}
	}
}

func handleChatWs(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// Just to make sure the server won't crash
		if r := recover(); r != nil {
			log.Println("recovered error in handleWs")
			if err, ok := r.(error); ok {
				log.Println(err.Error())
			} else {
				log.Println(r)
			}
		}
	}()

	var lobbyID, clientID string

	conn, err := chatUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}

	conn.SetCloseHandler(func(code int, text string) error {
		handleChatClose(lobbyID, clientID)
		return nil
	})

	messageType, msg, err := conn.ReadMessage()
	if err != nil {
		return
	}

	if messageType == websocket.TextMessage {
		token, err := jwt.Parse(string(msg), func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return sharedKey, nil
		})
		if err != nil {
			log.Println(err)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			clientID = claims["client_id"].(string)
			lobbyID = claims["lobby_id"].(string)
		} else {
			log.Println(err)
			return
		}
	}

	conn.WriteMessage(websocket.TextMessage, []byte("OK"))

	// If the lobby does not exist, make it
	if _, ok := chatLobbies[lobbyID]; !ok {
		chatLobbies[lobbyID] = make(map[string]*websocket.Conn)
	}

	lobby := chatLobbies[lobbyID]

	// If a user joins more than once, remove the previous instance
	if _, ok := lobby[clientID]; ok {
		handleChatClose(lobbyID, clientID)
	}

	lobby[clientID] = conn

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			handleChatClose(lobbyID, clientID)
			return
		}

		if messageType == websocket.TextMessage {
			for _, client := range lobby {
				client.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s:%s", clientID, msg)))
			}

			log.Printf("[chat] %s->%s: %s\n", clientID, lobbyID, msg)
		}
	}
}

var sharedKey []byte

func init() {
	sharedKey = make([]byte, 64)
	rand.Read(sharedKey)
}
