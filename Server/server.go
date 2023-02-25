package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gorilla/websocket"
	"golang.org/x/exp/slices"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Client struct {
	Server     *GameServer
	legalMoves []string
	closed     bool
	connMu     sync.Mutex // Protect conn
	conn       *websocket.Conn

	_lastSent []byte
}

func (c *Client) send(data []byte) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) Send(p *Packet) {
	data, err := json.Marshal(p)
	if err != nil {
		log.Println("Unmarsahal:", err)
		return
	}

	if p.Type == PacketTypeError {
		if err := c.send(data); err != nil {
			log.Println("Error sending error packet:", err)
			c.Disconnect()
		}

		return
	}

	patch, err := jsonpatch.CreateMergePatch(c._lastSent, data)
	if err != nil {
		log.Println("Error creating merge patch:", err)
		return
	}
	// If patch is empty, nothing needs to be sent
	if len(patch) == 2 {
		return
	}

	c._lastSent = data

	if err := c.send(patch); err != nil {
		log.Println("Error sending patch packet:", err)
		c.Disconnect()
	}
}

func (c *Client) SendError(err error) {
	c.Send(NewPacketError(err.Error()))
}

func (c *Client) Disconnect() {
	// Already closed once
	if c.closed {
		return
	}

	c.closed = true
	c.conn.Close()
	c.Server.game.Disconnect(c)
}

func (c *Client) Sync() {
	if c.closed {
		return
	}

	moves, data := c.Server.game.LegalMoves(c)
	if moves == nil {
		moves = []string{}
	}
	c.legalMoves = moves
	c.Send(&Packet{
		Type:  PacketTypeMessage,
		Moves: moves,
		Game:  c.Server.game.Name(c),
		State: c.Server.game.State(c),
		Data:  data,
	})
}

type GameServer struct {
	game Game
}

func NewGameServer(game Game) *GameServer {
	return &GameServer{
		game: game,
	}
}

func (s *GameServer) handleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}

	c := &Client{
		Server:     s,
		legalMoves: []string{},
		conn:       conn,
		_lastSent:  []byte("{}"),
	}
	c.Sync()

	conn.SetCloseHandler(func(code int, text string) error {
		c.Disconnect()
		return nil
	})

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			c.Disconnect()
			return
		}

		if messageType == websocket.TextMessage {
			packet := &Packet{}
			err := json.Unmarshal(msg, packet)
			if err != nil {
				log.Println("Unmarshal:", err)
				continue
			}

			if packet.Type != PacketTypeMessage {
				c.SendError(errors.New("unknown packet type"))
				continue
			}

			if len(packet.Moves) < 1 {
				c.SendError(errors.New("no moves sent"))
				continue
			}

			if len(packet.Moves) > 10 {
				c.SendError(errors.New("too many moves sent"))
				continue
			}

			illegalMoveMade := false
			for _, m := range packet.Moves {
				if !slices.Contains(c.legalMoves, m) {
					illegalMoveMade = true
					break
				}
			}
			if illegalMoveMade {
				c.SendError(errors.New("illegal move sent"))
				continue
			}

			var data map[string]interface{}
			if packet.Data != nil {
				data = packet.Data
			}

			if err := c.Server.game.ExecuteMoves(c, packet.Moves, data); err != nil {
				c.SendError(err)
				continue
			}
		}
	}
}
