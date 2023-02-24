package main

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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
		l.Sync()
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

func (lm *LobbyManager) disconnect(lobby *Lobby, client *Client) {
	if lobby.game == nil {
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
	} else {
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
	lobby := lm.Lobby(client)
	if lobby == nil {
		defer func() {
			// If the user is now in a lobby
			if lobby := lm.Lobby(client); lobby != nil {
				lobby.Sync()
			}
		}()
	} else {
		defer func() {
			lobby.Sync()
		}()
	}

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

		if lobby.game != nil {
			return errors.New("you cannot join a game in progress")
		}

		lobby.Clients[lc.ID] = lc
		lm.clientToLobby.Store(client, lobbyID)

	case MoveReconnect:
		lobbyID, _ := Get[string](data, "id")
		clientID, _ := Get[string](data, "me")

		entry, ok := lm.Lobbies.Load(lobbyID)
		if !ok {
			return errors.New("lobby no longer exists")
		}
		lobby := entry.(*Lobby)

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

	case MoveDisconnect:
		lm.disconnect(lobby, client)
		client.Sync()

	case MoveSelect:
		g, _ := Get[string](data, "game")

		if _, ok := GAMES[g]; !ok {
			return errors.New("invalid game")
		}

		lobby.Game = &g

	case MoveStart:
		g := GAMES[*lobby.Game]
		if g != nil {
			if len(lobby.Clients) < g.MinPlayers {
				return errors.New("too few players to start")
			}
			if len(lobby.Clients) > g.MaxPlayers {
				return errors.New("too many players to start")
			}
			lobby.game = g.Create(lobby)
		}

	case MoveRename:
		name, _ := Get[string](data, "name")
		name = cleanName(name)
		if name == "" {
			return errors.New("no name provided")
		}

		lobby.Client(client).Name = name

	case MoveKick:
		id, _ := Get[string](data, "id")

		if id == lobby.Client(client).ID {
			return errors.New("cannot kick yourself")
		}

		target, ok := lobby.Clients[id]
		if !ok {
			return errors.New("invalid ID provided")
		}

		target.SendError(errors.New("you have been kicked"))
		lm.disconnect(lobby, target.Client)
		target.Sync()

	case MoveTransfer:
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

	case MoveReturn:
		lobby.game = nil
		lobby.Frozen = false

		for id, c := range lobby.Clients {
			if c.Disconnected {
				delete(lobby.Clients, id)
			}
		}

	default:
		if lobby != nil && lobby.game != nil {
			return lobby.game.ExecuteMoves(client, moves, data)
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

	lm.disconnect(lobby, client)

	lobby.Sync()
}
