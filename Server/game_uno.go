package main

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

// This game consists of colored cards 0, 1, 2, ..., 9, reverse, skip, and +2 (rgby)
// and uncolored cards +4, change color.

// The # of milliseconds a player has to call uno for themself before anyone else can
const unoGracePeriod = 500

const (
	// moveDraw = "draw"
	moveUno = "uno"
)

type cardType int

const (
	number cardType = iota
	skip
	reverse
	drawTwo
	wild
	drawFour
)

type cardColor int

const (
	red cardColor = iota
	yellow
	blue
	green
	black
)

func parseCard(raw int) (t cardType, col cardColor, num int) {
	if raw >= 15*4 {
		// Skip 0s
		raw -= 14 * 4
	}

	num = -1
	col = cardColor(raw % 4)

	switch raw / 4 {
	case 10:
		t = skip
	case 11:
		t = reverse
	case 12:
		t = drawTwo
	case 13:
		t = wild
		col = black
	case 14:
		t = drawFour
		col = black
	default:
		t = number
		num = raw / 4
	}

	return
}

type Uno struct {
	Lobby         *Lobby    `json:"lobby"`
	CurrentPlayer int       `json:"current_player"` // Whose turn is it
	PlayPile      *Pile     `json:"play_pile"`      // What cards have been played so far
	PlayerOrder   []string  `json:"player_order"`   // What order do players play in
	Clockwise     bool      `json:"clockwise"`      // Are we moving clockwise
	DrawNum       int       `json:"draw_num"`       // The number of cards that are to be drawn by whoever accepts it (default 0)
	ChosenColor   cardColor `json:"chosen_color"`
	Winners       []string  `json:"winners"` // Who won and in what order

	unoAt    map[string]int64 // Whenever each player last reached one card
	drawPile *Pile
	hands    Hands
}

// Makes sure the draw pile has at least n cards. If draw pile could
// not be refilled, (i.e. players have drawn too many cards), this
// returns false
func (u *Uno) checkDrawPile(n int) bool {
	if len(*u.drawPile) < n {
		u.drawPile.Insert(u.PlayPile.Draw(len(*u.PlayPile) - 1))
		u.drawPile.Shuffle()
	}

	return len(*u.drawPile) >= n
}

type UnoState struct {
	Uno
	Hand       []int          `json:"hand"`
	OtherHands map[string]int `json:"other_hands"`
	Me         string         `json:"me"`
}

func NewUno(l *Lobby) FreezableGame {
	// One of every card in every color
	// + another 1-9 (no extra 0's), skip, reverse, d2
	p := CreatePile(15*4+12*4, false)
	p.Shuffle()

	g := &Uno{
		Lobby:       l,
		Clockwise:   true,
		ChosenColor: -1,
		Winners:     []string{},
		unoAt:       make(map[string]int64),
		drawPile:    p,
		hands:       Hands{},
	}

	for id := range l.Clients {
		cards := Pile(p.Draw(7))
		g.hands[id] = &cards
		g.PlayerOrder = append(g.PlayerOrder, id)
	}

	// First card must be a color card
	for {
		cards := Pile(p.Draw(1))
		_, col, _ := parseCard(cards[0])
		if col != black {
			g.PlayPile = &cards
			break
		}

		p.Insert(cards)
	}

	return g
}

func (g *Uno) isPlayable(next int) bool {
	nType, nCol, nNum := parseCard(next)
	// If the current player has to draw, they can only play +2 or +4
	if g.DrawNum > 0 {
		return nType == drawFour || nType == drawTwo
	}

	// If the card is black (+4/wild) it can always be played
	if nCol == black {
		return true
	}

	// Besides these conditions, we must compare with the last played card.
	lType, lCol, lNum := parseCard((*g.PlayPile)[len(*g.PlayPile)-1])

	// If the last card played changed the color, we must match the chosen color
	if lCol == black {
		return nCol == g.ChosenColor
	}

	// Same color can be played
	if lCol == nCol {
		return true
	}

	if nType == number {
		// Same number can be played
		return lNum == nNum
	}

	// Same type can be played
	return lType == nType
}

func (g *Uno) nextPlayer() {
	cur := g.CurrentPlayer
	for {
		if g.Clockwise {
			g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.PlayerOrder)
		} else {
			g.CurrentPlayer -= 1
			// Euclidean modulus
			if g.CurrentPlayer < 0 {
				g.CurrentPlayer += len(g.PlayerOrder)
			}
		}

		if g.CurrentPlayer == cur || len(*g.hands[g.PlayerOrder[g.CurrentPlayer]]) > 0 {
			break
		}
	}
}

func (g *Uno) ExecuteMoves(client *Client, moves []string, data interface{}) (err error) {
	lc := g.Lobby.Client(client)
	hand := g.hands[lc.ID]

	switch moves[0] {
	case moveDraw:
		num := 1
		if g.DrawNum > 0 {
			num = g.DrawNum
			g.nextPlayer()
			g.DrawNum = 0
		}

		enough := g.checkDrawPile(num)
		hand.Insert(g.drawPile.Draw(num))

		if !enough {
			return errors.New("all cards have been drawn")
		}

	case moveUno:
		for id, at := range g.unoAt {
			if at == 0 {
				continue
			}

			if id == lc.ID {
				g.unoAt[id] = 0
				continue
			}

			// If the grace period is over, force draw 2
			if at+unoGracePeriod < time.Now().UnixMilli() {
				g.checkDrawPile(2)
				g.hands[id].Insert(g.drawPile.Draw(2))
				g.unoAt[id] = 0
			}
		}

	default:
		split := strings.Split(moves[0], "_")
		// Get card id
		raw, _ := strconv.Atoi(split[0])
		// Add card to play pile
		g.PlayPile.Insert([]int{raw})
		// Remove card from player's hand
		i := slices.Index(*hand, raw)
		*hand = slices.Delete(*hand, i, i+1)
		// Clear previously chosen color
		g.ChosenColor = -1
		// Take special action depending on type of card
		ct, c, _ := parseCard(raw)
		switch ct {
		case reverse:
			g.Clockwise = !g.Clockwise
			// In a two player UNO game, the reverse card acts as a skip
			if len(g.PlayerOrder)-len(g.Winners) == 2 {
				g.nextPlayer()
			}
		case skip:
			g.nextPlayer()
		case drawTwo:
			g.DrawNum += 2
		case drawFour:
			g.DrawNum += 4
		}
		// Change color if we have a black card
		if c == black {
			col, _ := strconv.Atoi(split[1])
			g.ChosenColor = cardColor(col)
		}
		// Check if player has one card remaining
		if len(*hand) == 1 {
			g.unoAt[lc.ID] = time.Now().UnixMilli()
			g.Lobby.SyncAfter(unoGracePeriod * time.Millisecond)
		}
		// Check if player has won
		if len(*hand) == 0 {
			g.Winners = append(g.Winners, lc.ID)
			g.unoAt[lc.ID] = 0
		}
		// Move onto next player
		g.nextPlayer()
	}

	return
}

func (g *Uno) LegalMoves(client *Client) (moves []string, extra map[string]interface{}) {
	if len(g.Winners) >= 1 {
		moves = append(moves, MoveReturn)
	}

	c := g.Lobby.Client(client)
	for id, at := range g.unoAt {
		if at == 0 {
			continue
		}

		if id == c.ID {
			// Uno self
			moves = append(moves, moveUno)
		} else if at+unoGracePeriod < time.Now().UnixMilli() {
			// Uno someone else
			moves = append(moves, moveUno)
		}
	}
	if g.PlayerOrder[g.CurrentPlayer] != c.ID {
		return
	}

	moves = append(moves, moveDraw)

	for _, c := range *g.hands[c.ID] {
		if g.isPlayable(c) {
			_, col, _ := parseCard(c)
			s := strconv.Itoa(c)
			if col == black {
				for i := 0; i < int(black); i++ {
					moves = append(moves, s+"_"+strconv.Itoa(i))
				}
			} else {
				moves = append(moves, s)
			}
		}
	}

	return
}

// Name implements FreezableGame
func (*Uno) Name(client *Client) string {
	return "uno"
}

// State implements FreezableGame
func (g *Uno) State(client *Client) interface{} {
	c := g.Lobby.Client(client)
	return &UnoState{
		Uno:        *g,
		Hand:       *g.hands[c.ID],
		OtherHands: g.hands.StateHidden(c.ID),
		Me:         c.ID,
	}
}

func init() {
	registerGame(&GameData{
		Create:     NewUno,
		Name:       "uno",
		MinPlayers: 2,
		MaxPlayers: 4,
	})
}
