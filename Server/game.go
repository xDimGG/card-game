package main

import (
	"math/rand"

	"golang.org/x/exp/slices"
)

type FreezableGame interface {
	// Identifier of the game being played
	Name(client *Client) string
	// Returns the state of the game from the perspective of the given client
	State(client *Client) interface{}
	// Returns all legal moves along with any extra data about the moves
	LegalMoves(client *Client) ([]string, map[string]interface{})
	// Executes the given moves on the game
	ExecuteMoves(client *Client, moves []string, data interface{}) error
}

type Game interface {
	FreezableGame
	// Determines how players disconnecting should be handled
	Disconnect(client *Client)
}

type SmartGame interface {
	FreezableGame
	// A function to control bots playing this game
	SelectMoves(client *Client, moves []string) []string
}

type GameData struct {
	Create     func(*Lobby) FreezableGame
	Name       string
	MinPlayers int
	MaxPlayers int
}

var GAMES = make(map[string]*GameData)

func registerGame(g *GameData) {
	GAMES[g.Name] = g
}

type Pile []int

// Creates a new pile with the cards [1, n] if zero is false or [0, n) if zero is true
func CreatePile(n int, zero bool) *Pile {
	pile := Pile(make([]int, n))

	for i := 0; i < n; i++ {
		pile[i] = i
		if !zero {
			pile[i] += 1
		}
	}

	return &pile
}

func (p *Pile) Shuffle() {
	rand.Shuffle(len(*p), func(i, j int) {
		(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
	})
}

func (p *Pile) Draw(n int) []int {
	if n > len(*p) {
		n = len(*p)
	}

	drawn := (*p)[:n]
	*p = (*p)[n:]
	return slices.Clone(drawn)
}

// Insert cards into end of pile
func (p *Pile) Insert(cards []int) {
	*p = append(*p, cards...)
}

// Distributes entire pile to hands as evenly as possible
func (p *Pile) Distribute(h Hands) {
	n := len(*p) / len(h)
	extra := len(*p) % len(h)

	for id := range h {
		if extra > 0 {
			*h[id] = p.Draw(n + 1)
			extra -= 1
		} else {
			*h[id] = p.Draw(n)
		}
	}
}

type Hands map[string]*Pile

func (h *Hands) StateHidden(player string) map[string]int {
	hands := make(map[string]int)

	for id, cards := range *h {
		if id != player {
			hands[id] = len(*cards)
		}
	}

	return hands
}

func (h *Hands) StateRevealed(player string) Hands {
	hands := make(Hands)

	for id, cards := range *h {
		if id != player {
			hands[id] = cards
		}
	}

	return hands
}
