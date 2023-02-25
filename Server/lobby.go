package main

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sasha-s/go-deadlock"
)

const (
	MoveJoin       = "lobby.join"
	MoveReconnect  = "lobby.reconnect"
	MoveDisconnect = "lobby.disconnect"
	MoveStart      = "lobby.start"
	MoveSelect     = "lobby.select"
	MoveRename     = "lobby.rename"
	MoveKick       = "lobby.kick"
	MoveTransfer   = "lobby.transfer"
	MoveReturn     = "lobby.return"
)

type Lobby struct {
	Clients map[string]*LobbyClient `json:"clients"`
	Game    *string                 `json:"game"`
	ID      string                  `json:"id"`
	Frozen  bool                    `json:"frozen"`
	game    FreezableGame

	mu deadlock.RWMutex
}

type LobbyState struct {
	*Lobby
	Me string `json:"me"`
}

// Cleans newlines and removes extraneous spaces
func cleanName(name string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(name, "\n", " "), "\r", ""))
}

func (l *Lobby) Client(c *Client) *LobbyClient {
	for _, lc := range l.Clients {
		if lc.Client == c {
			return lc
		}
	}

	return nil
}

func (l *Lobby) Sync() {
	for _, c := range l.Clients {
		c.Client.Sync()
	}
}

func (l *Lobby) SyncAfter(t time.Duration) {
	go func() {
		time.Sleep(t)
		l.mu.Lock()
		l.Sync()
		l.mu.Unlock()
	}()
}

type LobbyManager struct {
	Lobbies       sync.Map
	clientToLobby sync.Map
}

type LobbyClient struct {
	*Client      `json:"-"`
	Name         string `json:"name"`
	Leader       bool   `json:"leader"`
	ID           string `json:"id"`
	JoinedAt     int64  `json:"joined_at"`
	Disconnected bool   `json:"disconnected"`
}

// Removes client from given lobby. Does not sync
func (lm *LobbyManager) remove(lobby *Lobby, client *Client) {
	// Remove client from lobby
	if c := lobby.Client(client); c != nil {
		delete(lobby.Clients, c.ID)
		lm.clientToLobby.Delete(client)

		if c.Leader && len(lobby.Clients) > 0 {
			var oldest *LobbyClient
			for _, nc := range lobby.Clients {
				if oldest == nil || nc.JoinedAt < oldest.JoinedAt {
					oldest = nc
				}
			}
			oldest.Leader = true
		}
	}

	// Delete lobby if empty
	if len(lobby.Clients) == 0 {
		lm.Lobbies.Delete(lobby.ID)
	}
}

func (lm *LobbyManager) Lobby(client *Client) *Lobby {
	if lobbyID, ok := lm.clientToLobby.Load(client); ok {
		if lobby, ok := lm.Lobbies.Load(lobbyID); ok {
			return lobby.(*Lobby)
		}
	}

	return nil
}

func (lm *LobbyManager) Name(client *Client) string {
	lobby := lm.Lobby(client)
	if lobby != nil {
		if lobby.game != nil {
			return lobby.game.Name(client)
		}
	}

	return "lobby"
}

func (lm *LobbyManager) LegalMoves(client *Client) (moves []string, extra map[string]interface{}) {
	lobby := lm.Lobby(client)

	// Do we have a new client?
	if lobby == nil {
		return []string{MoveJoin, MoveReconnect}, nil
	}

	if lobby.Frozen {
		if lobby.Client(client).Leader {
			moves = append(moves, MoveReturn)
		}

		return
	}

	if lobby.game != nil {
		moves, data := lobby.game.LegalMoves(client)
		if lobby.Client(client).Leader {
			moves = append(moves, MoveReturn)
		}

		return moves, data
	}

	if lobby.Client(client).Leader {
		// If a game has been selected
		if lobby.Game != nil {
			moves = append(moves, MoveStart)
		}

		return append(moves, MoveKick, MoveTransfer, MoveSelect, MoveRename, MoveDisconnect), nil
	}

	return []string{MoveRename, MoveDisconnect}, nil
}

func (lm *LobbyManager) ExecuteMoves(client *Client, moves []string, data interface{}) error {
	switch moves[0] {
	case MoveJoin:
		lobbyID, _ := Get[string](data, "lobby")
		name, _ := Get[string](data, "name")
		name = cleanName(name)

		if strings.TrimSpace(lobbyID) == "" {
			return errors.New("you must specify a lobby ID")
		}

		if name == "" {
			return errors.New("you must specify a name")
		}

		lc := &LobbyClient{
			Client:       client,
			Name:         name,
			Leader:       false,
			ID:           uuid.New().String(),
			JoinedAt:     time.Now().UnixMilli(),
			Disconnected: false,
		}

		entry, ok := lm.Lobbies.Load(lobbyID)
		var lobby *Lobby
		if ok {
			lobby = entry.(*Lobby)
		} else {
			lc.Leader = true
			lobby = &Lobby{
				ID:      lobbyID,
				Clients: make(map[string]*LobbyClient),
			}
			lm.Lobbies.Store(lobbyID, lobby)
		}

		lobby.mu.Lock()
		defer lobby.mu.Unlock()

		if lobby.game != nil {
			return errors.New("you cannot join a game in progress")
		}

		lobby.Clients[lc.ID] = lc
		lm.clientToLobby.Store(client, lobbyID)
		lobby.Sync()

	case MoveReconnect:
		lobbyID, _ := Get[string](data, "id")
		clientID, _ := Get[string](data, "me")

		entry, ok := lm.Lobbies.Load(lobbyID)
		if !ok {
			return errors.New("lobby no longer exists")
		}

		lobby := entry.(*Lobby)
		lobby.mu.Lock()
		defer lobby.mu.Unlock()

		if lobby.game == nil {
			return errors.New("game has ended")
		}

		lc, ok := lobby.Clients[clientID]
		if !ok {
			return errors.New("invalid client ID")
		}

		if !lc.Disconnected {
			return errors.New("you are already connected")
		}

		lc.Client = client
		lc.Disconnected = false
		lm.clientToLobby.Store(client, lobbyID)

		allConnected := true
		for _, c := range lobby.Clients {
			if c.Disconnected {
				allConnected = false
				break
			}
		}

		if allConnected {
			lobby.Frozen = false
		}

		lobby.Sync()

	case MoveDisconnect:
		lobby := lm.Lobby(client)
		lobby.mu.Lock()
		lm.remove(lobby, client)
		lobby.Sync()
		lobby.mu.Unlock()
		client.Sync()

	case MoveSelect:
		g, _ := Get[string](data, "game")

		if _, ok := GAMES[g]; !ok {
			return errors.New("invalid game")
		}

		lobby := lm.Lobby(client)
		lobby.mu.Lock()
		lobby.Game = &g
		lobby.Sync()
		lobby.mu.Unlock()

	case MoveStart:
		lobby := lm.Lobby(client)
		lobby.mu.Lock()
		defer lobby.mu.Unlock()

		g := GAMES[*lobby.Game]
		if len(lobby.Clients) < g.MinPlayers {
			return errors.New("too few players to start")
		}
		if len(lobby.Clients) > g.MaxPlayers {
			return errors.New("too many players to start")
		}

		lobby.game = g.Create(lobby)
		lobby.Sync()

	case MoveRename:
		name, _ := Get[string](data, "name")
		name = cleanName(name)
		if name == "" {
			return errors.New("no name provided")
		}

		lobby := lm.Lobby(client)
		lobby.mu.Lock()
		lobby.Client(client).Name = name
		lobby.Sync()
		lobby.mu.Unlock()

	case MoveKick:
		lobby := lm.Lobby(client)
		lobby.mu.Lock()
		defer lobby.mu.Unlock()

		id, _ := Get[string](data, "id")

		if id == lobby.Client(client).ID {
			return errors.New("cannot kick yourself")
		}

		target, ok := lobby.Clients[id]
		if !ok {
			return errors.New("invalid ID provided")
		}

		lm.remove(lobby, target.Client)
		lobby.Sync()

		target.SendError(errors.New("you have been kicked"))
		target.Sync()

	case MoveTransfer:
		lobby := lm.Lobby(client)
		lobby.mu.Lock()
		defer lobby.mu.Unlock()

		id, _ := Get[string](data, "id")

		target, ok := lobby.Clients[id]
		if !ok {
			return errors.New("invalid ID provided")
		}

		if target.Leader {
			return errors.New("client is already leader")
		}

		lobby.Client(client).Leader = false
		target.Leader = true
		lobby.Sync()

	case MoveReturn:
		lobby := lm.Lobby(client)
		lobby.mu.Lock()
		defer lobby.mu.Unlock()
		lobby.game = nil
		lobby.Frozen = false

		for id, c := range lobby.Clients {
			if c.Disconnected {
				delete(lobby.Clients, id)
			}
		}

		lobby.Sync()

	default:
		lobby := lm.Lobby(client)
		if lobby != nil {
			lobby.mu.Lock()
			defer func() {
				lobby.mu.Unlock()
				lobby.Sync()
			}()

			if lobby.game != nil {
				return lobby.game.ExecuteMoves(client, moves, data)
			}
		}
	}

	return nil
}

func (lm *LobbyManager) State(client *Client) interface{} {
	lobby := lm.Lobby(client)
	if lobby == nil {
		return nil
	}

	if lobby.game != nil {
		return lobby.game.State(client)
	}

	// Maybe copy lobby
	return LobbyState{
		Lobby: lobby,
		Me:    lobby.Client(client).ID,
	}
}

func (lm *LobbyManager) Disconnect(client *Client) {
	lobby := lm.Lobby(client)
	if lobby == nil {
		return
	}

	lobby.mu.Lock()
	defer func() {
		lobby.Sync()
		lobby.mu.Unlock()
	}()

	if lobby.game == nil {
		lm.remove(lobby, client)
		return
	}

	if g, ok := lobby.game.(Game); ok {
		g.Disconnect(client)
		return
	}

	lobby.Frozen = true
	lobby.Client(client).Disconnected = true

	// Delete lobby if everyone is disconnected
	someone := false
	for _, c := range lobby.Clients {
		if !c.Disconnected {
			someone = true
			break
		}
	}

	if !someone {
		lm.Lobbies.Delete(lobby.ID)
	}
}