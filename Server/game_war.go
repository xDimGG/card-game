package main

import (
	"strconv"

	"golang.org/x/exp/slices"
)

const (
	moveWar = "war"
	// moveNextRound = "next_round"
)

type WarPhase int

const (
	WarPhasePreparation WarPhase = iota
	WarPhaseReady
	WarPhaseReveal
)

type War struct {
	Wins         map[string]int `json:"wins"`
	Lobby        *Lobby         `json:"lobby"`
	RoundHighest int            `json:"round_highest,omitempty"`

	phase  WarPhase
	placed map[string]int
	hands  Hands
}

type WarState struct {
	War
	Placed         int             `json:"placed,omitempty"`
	OtherPlaced    map[string]bool `json:"other_placed"`
	PlacedRevealed map[string]int  `json:"placed_revealed"`
	Hand           []int           `json:"hand"`
	OtherHands     map[string]int  `json:"other_hands"`
}

func NewWar(l *Lobby) FreezableGame {
	g := &War{Lobby: l}
	pile := CreatePile(len(g.Lobby.Clients)*10, false)
	pile.Shuffle()

	g.phase = WarPhasePreparation
	g.hands = Hands{}
	g.placed = make(map[string]int)
	g.Wins = make(map[string]int)
	i := 0
	for id := range g.Lobby.Clients {
		p := Pile(pile.Draw(10))
		g.hands[id] = &p
		g.Wins[id] = 0
		i += 1
	}
	return g
}

func (game *War) ExecuteMoves(client *Client, moves []string, data interface{}) error {
	c := game.Lobby.Client(client)

	switch moves[0] {
	case moveWar:
		game.phase = WarPhaseReveal
		highest, winner := 0, ""
		for id, num := range game.placed {
			if num > highest {
				highest = num
				winner = id
			}
		}

		game.RoundHighest = highest
		game.Wins[winner] += 1

	case moveNextRound:
		game.placed = make(map[string]int)
		game.phase = WarPhasePreparation
		game.RoundHighest = 0

	default:
		input, err := strconv.Atoi(moves[0])
		if err != nil {
			return err
		}

		hand := game.hands[c.ID]
		i := slices.Index(*hand, input)
		*game.hands[c.ID] = slices.Delete(*game.hands[c.ID], i, i+1)
		game.placed[c.ID] = input

		allPlaced := true
		for id := range game.Lobby.Clients {
			if _, ok := game.placed[id]; !ok {
				allPlaced = false
				break
			}
		}

		if allPlaced {
			game.phase = WarPhaseReady
		}
	}

	return nil
}

func (game *War) LegalMoves(client *Client) ([]string, map[string]interface{}) {
	if game.phase == WarPhaseReady {
		return []string{moveWar}, nil
	}

	if game.phase == WarPhaseReveal {
		return []string{moveNextRound}, nil
	}

	c := game.Lobby.Client(client)
	if _, ok := game.placed[c.ID]; ok {
		return nil, nil
	}

	hand := game.hands[c.ID]
	moves := make([]string, len(*hand))
	for i, num := range *hand {
		moves[i] = strconv.Itoa(num)
	}

	return moves, nil
}

func (*War) Name(client *Client) string {
	return "war"
}

func (game *War) State(client *Client) interface{} {
	c := game.Lobby.Client(client)
	ws := &WarState{
		War:        *game,
		Placed:     game.placed[c.ID],
		Hand:       *game.hands[c.ID],
		OtherHands: game.hands.StateHidden(c.ID),
	}

	if game.phase == WarPhasePreparation || game.phase == WarPhaseReady {
		ws.OtherPlaced = map[string]bool{}

		for id := range game.Lobby.Clients {
			if c.ID != id {
				_, exists := game.placed[id]
				ws.OtherPlaced[id] = exists
			}
		}
	} else if game.phase == WarPhaseReveal {
		ws.PlacedRevealed = make(map[string]int)
		for id := range game.Lobby.Clients {
			if c.ID != id {
				ws.PlacedRevealed[id] = game.placed[id]
			}
		}
	}

	return ws
}

func init() {
	registerGame(&GameData{
		Create:     NewWar,
		Name:       "war",
		MinPlayers: 2,
		MaxPlayers: 4,
	})
}
