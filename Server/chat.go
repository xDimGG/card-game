package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var chatUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// map of chatLobbies to clients to connections
// var chatLobbies = map[string]map[string]*websocket.Conn{}
var chatLobbies sync.Map

func handleChatClose(lobbyID string, clientID string) {
	if lobbyAny, ok := chatLobbies.Load(lobbyID); ok {
		lobby := lobbyAny.(*sync.Map)
		if clientAny, ok := lobby.Load(clientID); ok {
			client := clientAny.(*websocket.Conn)
			client.Close()
			lobby.Delete(clientID)
		}

		// Remove empty lobbies from map
		empty := true
		lobby.Range(func(_, _ interface{}) bool {
			empty = false
			return false
		})

		if empty {
			chatLobbies.Delete(lobbyID)
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

	lobbyAny, _ := chatLobbies.LoadOrStore(lobbyID, &sync.Map{})
	lobby := lobbyAny.(*sync.Map)

	// If a user joins more than once, remove the previous instance
	if _, ok := lobby.Load(clientID); ok {
		handleChatClose(lobbyID, clientID)
	}

	lobby.Store(clientID, conn)

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			handleChatClose(lobbyID, clientID)
			return
		}

		if messageType == websocket.TextMessage {
			lobby.Range(func(_, clientAny interface{}) bool {
				client := clientAny.(*websocket.Conn)
				client.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s:%s", clientID, msg)))

				return true
			})

			log.Printf("[chat] %s->%s: %s\n", clientID, lobbyID, msg)
		}
	}
}
