package main

import (
	"math"
	"math/rand"
	"slices"
)

// GameState provides context for AI decision making
type GameState struct {
	Round         int
	Players       []PlayerInterface
	ActivePlayers []PlayerInterface
	CurrentLeader PlayerInterface
	CardsInDeck   []*Card
}

type HitOrStayStrategy func(self PlayerInterface, gameState *GameState) bool
type ActionTargetStrategy func(self PlayerInterface, gameState *GameState, actionType ActionType) PlayerInterface

type ComputerPlayer struct {
	BasePlayer
	HitOrStayStrategy            HitOrStayStrategy
	ActionTargetStrategy         ActionTargetStrategy
	PositiveActionTargetStrategy ActionTargetStrategy
}

// NewComputerPlayer creates a new computer player with specified strategy
func NewComputerPlayer(name string, strategy HitOrStayStrategy, actionTargetStrategy ActionTargetStrategy, positiveActionTargetStrategy ActionTargetStrategy) *ComputerPlayer {
	p := &ComputerPlayer{
		HitOrStayStrategy:            strategy,
		ActionTargetStrategy:         actionTargetStrategy,
		PositiveActionTargetStrategy: positiveActionTargetStrategy,
	}

	p.BasePlayer.Init(name)

	return p
}

func (p *ComputerPlayer) GetPlayerIcon() string {
	return "🤖"
}

func (p *ComputerPlayer) MakeHitStayDecision(gameState *GameState) (bool, error) {
	// Always hit if you have a second chance
	if p.HasSecondChance() {
		return true, nil
	}

	return p.HitOrStayStrategy(p, gameState), nil
}

func (p *ComputerPlayer) ChooseActionTarget(gameState *GameState, actionType ActionType) (PlayerInterface, error) {
	return p.ActionTargetStrategy(p, gameState, actionType), nil
}

func (p *ComputerPlayer) ChoosePositiveActionTarget(gameState *GameState, actionType ActionType) (PlayerInterface, error) {
	return p.PositiveActionTargetStrategy(p, gameState, actionType), nil
}

func PlayRoundTo(n int) HitOrStayStrategy {
	return func(self PlayerInterface, gameState *GameState) bool {
		return self.CalculateRoundScore() < n
	}
}

func PlayToBustProbability(p float64) HitOrStayStrategy {
	return func(self PlayerInterface, gameState *GameState) bool {
		nBust := 0
		nTotal := len(gameState.CardsInDeck)
		numberCards := make([]int, 0)
		for _, card := range self.GetHand() {
			if card.Type == NumberCard {
				numberCards = append(numberCards, card.Value)
			}
		}

		for _, possibleCard := range gameState.CardsInDeck {
			if possibleCard.Type == NumberCard {
				if slices.Contains(numberCards, possibleCard.Value) {
					nBust++
				}
			}
		}

		return float64(nBust)/float64(nTotal) < p
	}
}

func HitUntilAheadBy(n int) HitOrStayStrategy {
	return func(self PlayerInterface, gameState *GameState) bool {
		return gameState.CurrentLeader.GetTotalScore()+gameState.CurrentLeader.CalculateRoundScore() < self.GetTotalScore()+self.CalculateRoundScore()+n
	}
}

func AlwaysHitStrategy(self PlayerInterface, gameState *GameState) bool {
	return true
}

func RandomHitOrStayStrategy(self PlayerInterface, gameState *GameState) bool {
	return rand.Intn(2) == 0
}

func TargetLeaderStrategy(self PlayerInterface, gameState *GameState, actionType ActionType) PlayerInterface {
	var leader PlayerInterface
	leaderScore := 0
	for _, player := range gameState.ActivePlayers {
		if player != self {
			playerScore := player.GetTotalScore() + player.CalculateRoundScore()
			if playerScore > leaderScore {
				leader = player
				leaderScore = playerScore
			}
		}
	}

	// Must target self if no other player is active
	if leader == nil {
		return self
	}

	return leader
}

func TargetLastPlaceStrategy(self PlayerInterface, gameState *GameState, actionType ActionType) PlayerInterface {
	var last PlayerInterface
	lastScore := math.MaxInt
	for _, player := range gameState.ActivePlayers {
		if player != self {
			playerScore := player.GetTotalScore() + player.CalculateRoundScore()
			if playerScore < lastScore {
				last = player
				lastScore = playerScore
			}
		}
	}

	// Must target self if no other player is active
	if last == nil {
		return self
	}

	return last
}

func TargetRandomStrategy(self PlayerInterface, gameState *GameState, actionType ActionType) PlayerInterface {
	activePlayers := make([]PlayerInterface, 0)
	for _, player := range gameState.Players {
		if player.IsActive() && player != self {
			activePlayers = append(activePlayers, player)
		}
	}

	// Must target self if no other player is active
	if len(activePlayers) == 0 {
		return self
	}

	return activePlayers[rand.Intn(len(activePlayers))]
}
