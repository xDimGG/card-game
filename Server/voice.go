package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var voiceUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// map of voiceLobbies to clients to connections
// var voiceLobbies = map[string]map[string]*websocket.Conn{}
var voiceLobbies sync.Map

func handlevoiceClose(lobbyID string, clientID string) {
	if lobbyAny, ok := voiceLobbies.Load(lobbyID); ok {
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
			voiceLobbies.Delete(lobbyID)
		}
	}
}

func handleVoiceWs(w http.ResponseWriter, r *http.Request) {
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

	conn, err := voiceUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}

	conn.SetCloseHandler(func(code int, text string) error {
		handlevoiceClose(lobbyID, clientID)
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

	lobbyAny, _ := voiceLobbies.LoadOrStore(lobbyID, &sync.Map{})
	lobby := lobbyAny.(*sync.Map)

	// If a user joins more than once, remove the previous instance
	if _, ok := lobby.Load(clientID); ok {
		handlevoiceClose(lobbyID, clientID)
	}

	lobby.Range(func(_, clientAny interface{}) bool {
		client := clientAny.(*websocket.Conn)
		client.WriteMessage(websocket.TextMessage, []byte(clientID))

		return true
	})

	lobby.Store(clientID, conn)

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			handlevoiceClose(lobbyID, clientID)
			return
		}

		if messageType == websocket.TextMessage {
			log.Printf("[voice] %s->%s: %s\n", clientID, lobbyID, msg)

			if len(msg) < 40 {
				continue
			}
			recipient := string(msg[:len(clientID)])
			if recipient == clientID {
				continue
			}

			// Replace recipient ID with sender ID
			copy(msg, clientID)

			if clientAny, ok := lobby.Load(recipient); ok {
				receiver := clientAny.(*websocket.Conn)
				receiver.WriteMessage(websocket.TextMessage, msg)
			}
		}
	}
}
