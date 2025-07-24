package main

import "math/rand"

// GameState provides context for AI decision making
type GameState struct {
	Round         int
	Players       []PlayerInterface
	ActivePlayers []PlayerInterface
	CurrentLeader PlayerInterface
	CardsLeft     int
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

func AlwaysHitStrategy(self PlayerInterface, gameState *GameState) bool {
	return true
}

func RandomHitOrStayStrategy(self PlayerInterface, gameState *GameState) bool {
	return rand.Intn(2) == 0
}

func TargetLeaderStrategy(self PlayerInterface, gameState *GameState, actionType ActionType) PlayerInterface {
	var leader PlayerInterface
	for _, player := range gameState.Players {
		if player.IsActive() && player != self {
			if leader == nil || player.GetTotalScore()+player.CalculateRoundScore() > leader.GetTotalScore()+leader.CalculateRoundScore() {
				leader = player
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
	for _, player := range gameState.Players {
		if player.IsActive() && player != self {
			if last == nil || player.GetTotalScore()+player.CalculateRoundScore() < last.GetTotalScore()+last.CalculateRoundScore() {
				last = player
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
