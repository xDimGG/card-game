package main

import (
	"strconv"

	"golang.org/x/exp/slices"
)

const (
	moveUseShuriken = "use_shuriken"
	moveNextRound   = "next_round"
	moveRetryRound  = "retry_round"
	moveRestartGame = "restart_game"
)

type Reward string

const (
	RewardShuriken Reward = "shuriken"
	RewardLife     Reward = "life"
	RewardNone     Reward = ""
)

func getReward(round int) Reward {
	// round 2 reward is 1 shuriken, 3 is 1 life, 5 is 1 shuriken, etc.
	rewardRounds := []int{2, 3, 5, 6, 8, 9}
	i := -1
	for j, num := range rewardRounds {
		if num == round {
			i = j
			break
		}
	}

	if i == -1 {
		return RewardNone
	}

	if i%2 == 0 {
		return RewardShuriken
	}

	return RewardLife
}

type TheMindRound struct {
	Won         bool           `json:"won"`
	Lost        bool           `json:"lost"`
	PlayPile    []int          `json:"play_pile"`
	LowestCards map[string]int `json:"lowest_cards,omitempty"`
}

type TheMind struct {
	Round     TheMindRound `json:"round"`
	RoundNum  int          `json:"round_num"`
	Shurikens int          `json:"shurikens"`
	Lives     int          `json:"lives"`
	Won       bool         `json:"won"`
	Lost      bool         `json:"lost"`
	Lobby     *Lobby       `json:"lobby"`

	hands    Hands
	drawPile *Pile
}

type TheMindState struct {
	TheMind
	Hand              []int          `json:"hand"`
	OtherHands        map[string]int `json:"other_hands"`
	OtherHandsExposed Hands          `json:"other_hands_exposed"`
}

func NewTheMind(lobby *Lobby) FreezableGame {
	game := &TheMind{Lobby: lobby}
	game.Initialize()
	return game
}

func (game *TheMind) Initialize() {
	game.Round = TheMindRound{}
	game.RoundNum = 1
	game.Shurikens = 1
	game.Lives = len(game.Lobby.Clients)
	game.Won = false
	game.Lost = false
	game.BeginRound()
}

func (game *TheMind) GeneratePile() {
	game.drawPile = CreatePile(100, false)
	game.drawPile.Shuffle()
}

func (game *TheMind) BeginRound() {
	game.GeneratePile()
	game.Round.PlayPile = []int{}
	game.hands = Hands{}
	for _, client := range game.Lobby.Clients {
		p := Pile(game.drawPile.Draw(game.RoundNum))
		game.hands[client.ID] = &p
	}
}

func (game *TheMind) AdvanceRound() {
	switch getReward(game.RoundNum) {
	case RewardShuriken:
		game.Shurikens += 1
	case RewardLife:
		game.Lives += 1
	}

	game.drawPile.Insert(game.Round.PlayPile)
	game.drawPile.Shuffle()

	game.Round = TheMindRound{}
	game.RoundNum += 1
	game.BeginRound()
}

func (game *TheMind) RetryRound() {
	game.Round = TheMindRound{}
	game.BeginRound()
}

func (game *TheMind) Name(_ *Client) string {
	return "the_mind"
}

func (game *TheMind) State(client *Client) interface{} {
	lc := game.Lobby.Client(client)

	s := &TheMindState{
		TheMind: *game,
		Hand:    *game.hands[lc.ID],
	}

	if game.Round.Lost || game.Round.Won {
		s.OtherHandsExposed = game.hands.StateRevealed(lc.ID)
	} else {
		s.OtherHands = game.hands.StateHidden(lc.ID)
	}

	return s
}

func (game *TheMind) LegalMoves(client *Client) ([]string, map[string]interface{}) {
	if game.Won || game.Lost {
		return []string{moveRestartGame}, nil
	}

	if game.Round.Won {
		return []string{moveNextRound}, map[string]interface{}{
			"reward": getReward(game.RoundNum),
		}
	}

	if game.Round.Lost {
		return []string{moveRetryRound}, nil
	}

	lc := game.Lobby.Client(client)
	cards := game.hands[lc.ID]
	moves := make([]string, len(*cards))
	for i, card := range *cards {
		moves[i] = strconv.Itoa(card)
	}

	if game.Shurikens > 0 {
		moves = append(moves, moveUseShuriken)
	}

	return moves, nil
}

func (game *TheMind) ExecuteMoves(client *Client, moves []string, data interface{}) error {
	switch moves[0] {
	case moveUseShuriken:
		game.Shurikens -= 1
		game.Round.LowestCards = make(map[string]int)

		for id, hand := range game.hands {
			lowest := 101
			for _, el := range *hand {
				if el < lowest {
					lowest = el
				}
			}

			if lowest <= 100 {
				game.Round.LowestCards[id] = lowest
			}
		}

	case moveNextRound:
		game.AdvanceRound()

	case moveRetryRound:
		game.RetryRound()

	case moveRestartGame:
		game.Initialize()

	default:
		input, err := strconv.Atoi(moves[0])
		if err != nil {
			return err
		}

		lc := game.Lobby.Client(client)
		hand := game.hands[lc.ID]
		i := slices.Index(*hand, input)
		*game.hands[lc.ID] = slices.Delete(*game.hands[lc.ID], i, i+1)

		if game.Round.LowestCards != nil {
			c := game.Lobby.Client(client)
			if c != nil {
				// Checking if for some reason this player doesn't play their lowest card
				if game.Round.LowestCards[c.ID] == input {
					delete(game.Round.LowestCards, c.ID)
				}
			}
		}

		game.Round.PlayPile = append(game.Round.PlayPile, input)

		lastCard := game.Round.PlayPile[len(game.Round.PlayPile)-1]

		for _, hand := range game.hands {
			for _, card := range *hand {
				if card < lastCard {
					game.Round.Lost = true
					game.Lives -= 1
					if game.Lives == 0 {
						game.Lost = true
					}
					return nil
				}
			}
		}

		if len(game.Round.PlayPile) == game.RoundNum*len(game.Lobby.Clients) {
			game.Round.Won = true
			// 2 players: 12 rounds to win
			// 3 players: 10 rounds to win
			// 4 players: 8 rounds to win
			if game.RoundNum >= 14-(len(game.Lobby.Clients)*2) {
				game.Won = true
			}
		}
	}

	return nil
}

func init() {
	registerGame(&GameData{
		Create:     NewTheMind,
		Name:       "the_mind",
		MinPlayers: 2,
		MaxPlayers: 4,
	})
}
