package main

import "golang.org/x/exp/slices"

// 56 cards, 4 fruits (14 cards per fruit)
// 0-13: Strawberry, 14-27: Plum, 28-41: Pear, 42-55: Banana
// fruit(card) = floor(card / 14)
// Card mod 14:
// 0-2: 1 fruit
// 3-5: 2 fruits
// 6-8: 3 fruits
// 9-11:  4 fruits
// 12-13: 5 fruits
// fruits(card) = floor((card % 14) / 3) + 1

const (
	moveDraw  = "draw"
	movePress = "press"
)

type HG struct {
	Out           map[string]bool `json:"out"`            // Which players are out
	CurrentPlayer int             `json:"current_player"` // Whose turn is it
	PlayerOrder   []string        `json:"player_order"`   // What order do players play in
	PlayedCards   Hands           `json:"played_cards"`   // What cards have been played so far
	Winner        string          `json:"winner,omitempty"`

	lobby *Lobby
	hands Hands
}

type HGState struct {
	HG

	Hands map[string]int `json:"hands"` // All player hands
	Lobby *LobbyState    `json:"lobby"`
}

func NewHG(l *Lobby) FreezableGame {
	g := &HG{lobby: l}
	pile := CreatePile(56, true)
	pile.Shuffle()

	g.hands = Hands{}
	g.Out = make(map[string]bool)
	g.PlayedCards = Hands{}
	g.PlayerOrder = []string{}
	for id := range g.lobby.Clients {
		g.hands[id] = &Pile{}
		g.PlayedCards[id] = &Pile{}
		g.Out[id] = false
		g.PlayerOrder = append(g.PlayerOrder, id)
	}

	pile.Distribute(g.hands)
	g.CurrentPlayer = 0
	return g
}

// Choose the next person with cards
func (game *HG) Advance() {
	for {
		game.CurrentPlayer++
		game.CurrentPlayer %= len(game.PlayerOrder)

		if len(*game.hands[game.PlayerOrder[game.CurrentPlayer]]) > 0 {
			break
		}
	}
}

func (game *HG) ExecuteMoves(client *Client, moves []string, data interface{}) error {
	c := game.lobby.Client(client)

	switch moves[0] {
	case moveDraw:
		h := game.hands[c.ID]
		p := game.PlayedCards[c.ID]
		p.Insert(h.Draw(1))
		game.Advance()

	case movePress:
		fruits := []int{0, 0, 0, 0}
		for _, pile := range game.PlayedCards {
			if len(*pile) == 0 {
				continue
			}
			card := (*pile)[len(*pile)-1]
			fruits[card/14] += (card%14)/3 + 1
		}

		h := game.hands[c.ID]
		if slices.Contains(fruits, 5) {
			for id, p := range game.PlayedCards {
				h.Insert(*p)
				game.PlayedCards[id] = &Pile{}
			}

			h.Shuffle()
			game.CurrentPlayer = slices.Index(game.PlayerOrder, c.ID)

			nOut := 0
			winner := ""
			for id, cards := range game.hands {
				if len(*cards) == 0 {
					nOut += 1
					game.Out[id] = true
				} else {
					winner = id
				}
			}

			// If all but 1 players is out
			if nOut == len(game.lobby.Clients)-1 {
				game.Winner = winner
			}
		} else {
			for id := range game.lobby.Clients {
				if !game.Out[id] && id != c.ID {
					o := game.hands[id]
					o.Insert(h.Draw(1))
				}
			}

			if game.PlayerOrder[game.CurrentPlayer] == c.ID && len(*game.hands[c.ID]) == 0 {
				game.Advance()
			}
		}
	}

	return nil
}

func (game *HG) LegalMoves(client *Client) (moves []string, _ map[string]interface{}) {
	c := game.lobby.Client(client)

	if game.Winner != "" {
		return []string{MoveReturn}, nil
	}

	if game.Out[c.ID] {
		return
	}

	// Must be a player's turn to draw
	if c.ID == game.PlayerOrder[game.CurrentPlayer] && len(*game.hands[c.ID]) > 0 {
		moves = append(moves, moveDraw)
	}

	// The bell can only be pressed if at least one card has been played
	for _, pile := range game.PlayedCards {
		if len(*pile) > 0 {
			moves = append(moves, movePress)
			break
		}
	}

	return
}

func (*HG) Name(client *Client) string {
	return "halli_galli"
}

func (game *HG) State(client *Client) interface{} {
	return &HGState{
		HG:    *game,
		Lobby: game.lobby.State(client),
		Hands: game.hands.StateHidden(""),
	}
}

func init() {
	registerGame(&GameData{
		Create:     NewHG,
		Name:       "halli_galli",
		MinPlayers: 2,
		MaxPlayers: 4,
	})
}
