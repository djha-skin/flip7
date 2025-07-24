package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// PlayerState represents the current state of a player in a round
type PlayerState int

const (
	Active PlayerState = iota
	Stayed
	Busted
)

// PlayerType represents whether a player is human or computer
type PlayerType int

const (
	Human PlayerType = iota
	Computer
)

// AIStrategy represents different AI playing strategies
type AIStrategy int

const (
	Conservative AIStrategy = iota // Plays it safe, stays early
	Aggressive                     // Pushes luck for high scores
	Adaptive                       // Changes based on game state
	Chaotic                        // Unpredictable/fun decisions
)

// Player represents a game player
type Player struct {
	Name            string
	PlayerType      PlayerType
	AIStrategy      AIStrategy
	TotalScore      int
	RoundScore      int
	NumberCards     []*Card
	ModifierCards   []*Card
	ActionCards     []*Card
	State           PlayerState
	HasSecondChance bool
}

// NewHumanPlayer creates a new human player
func NewHumanPlayer(name string, scanner *bufio.Scanner) *Player {
	return &Player{
		Name:            name,
		PlayerType:      Human,
		AIStrategy:      Conservative, // Not used for human players
		TotalScore:      0,
		RoundScore:      0,
		NumberCards:     make([]*Card, 0),
		ModifierCards:   make([]*Card, 0),
		ActionCards:     make([]*Card, 0),
		State:           Active,
		HasSecondChance: false,
	}
}

// NewComputerPlayer creates a new computer player with specified strategy
func NewComputerPlayer(name string, strategy AIStrategy) *Player {
	return &Player{
		Name:            name,
		PlayerType:      Computer,
		AIStrategy:      strategy,
		TotalScore:      0,
		RoundScore:      0,
		NumberCards:     make([]*Card, 0),
		ModifierCards:   make([]*Card, 0),
		ActionCards:     make([]*Card, 0),
		State:           Active,
		HasSecondChance: false,
	}
}

// IsComputer returns true if this is a computer player (kept for backwards compatibility)
func (p *Player) IsComputer() bool {
	return p.PlayerType == Computer
}

// IsHuman returns true if this is a human player (kept for backwards compatibility)
func (p *Player) IsHuman() bool {
	return p.PlayerType == Human
}

// AddCard adds a card to the player's hand
func (p *Player) AddCard(card *Card) error {
	switch card.Type {
	case NumberCard:
		// Check for duplicate number cards (busting)
		for _, existing := range p.NumberCards {
			if existing.Value == card.Value {
				// Player busts unless they have a second chance
				if p.HasSecondChance {
					return fmt.Errorf("duplicate_with_second_chance:%d", card.Value)
				}
				p.State = Busted
				return fmt.Errorf("bust:%d", card.Value)
			}
		}
		p.NumberCards = append(p.NumberCards, card)

		// Check for Flip 7
		if len(p.NumberCards) == 7 {
			p.State = Stayed // Player automatically stays when achieving Flip 7
			return fmt.Errorf("flip7")
		}

	case ModifierCard:
		p.ModifierCards = append(p.ModifierCards, card)

	case ActionCard:
		if card.Action == SecondChance {
			if p.HasSecondChance {
				return fmt.Errorf("second_chance_duplicate")
			}
			p.HasSecondChance = true
		}
		p.ActionCards = append(p.ActionCards, card)
	}

	return nil
}

// UseSecondChance uses the second chance card to avoid busting
func (p *Player) UseSecondChance(duplicateValue int) {
	if !p.HasSecondChance {
		return
	}

	// Remove the duplicate number card
	for i, card := range p.NumberCards {
		if card.Value == duplicateValue {
			p.NumberCards = append(p.NumberCards[:i], p.NumberCards[i+1:]...)
			break
		}
	}

	// Remove second chance card
	for i, card := range p.ActionCards {
		if card.Action == SecondChance {
			p.ActionCards = append(p.ActionCards[:i], p.ActionCards[i+1:]...)
			break
		}
	}

	p.HasSecondChance = false
}

// Stay makes the player stay and bank their points
func (p *Player) Stay() {
	if p.State == Active {
		p.State = Stayed
	}
}

// CalculateRoundScore calculates the player's score for the current round
func (p *Player) CalculateRoundScore() int {
	if p.State == Busted {
		p.RoundScore = 0
		return 0
	}

	// Calculate base score from number cards
	numberTotal := 0
	for _, card := range p.NumberCards {
		numberTotal += card.Value
	}

	// Apply multiplier if present
	for _, card := range p.ModifierCards {
		if card.Modifier == Multiply2 {
			numberTotal *= 2
			break
		}
	}

	// Add modifier points
	modifierTotal := 0
	for _, card := range p.ModifierCards {
		if card.Modifier != Multiply2 {
			modifierTotal += card.GetPoints()
		}
	}

	total := numberTotal + modifierTotal

	// Add Flip 7 bonus
	if len(p.NumberCards) == 7 {
		total += 15
	}

	p.RoundScore = total
	return total
}

// AddToTotalScore adds the round score to the total score
func (p *Player) AddToTotalScore() {
	p.TotalScore += p.RoundScore
}

// ResetForNewRound resets the player's state for a new round
func (p *Player) ResetForNewRound() {
	p.RoundScore = 0
	p.NumberCards = make([]*Card, 0)
	p.ModifierCards = make([]*Card, 0)
	p.ActionCards = make([]*Card, 0)
	p.State = Active
	p.HasSecondChance = false
}

// IsActive returns true if the player is still active in the current round
func (p *Player) IsActive() bool {
	return p.State == Active
}

// HasCards returns true if the player has any number cards
func (p *Player) HasCards() bool {
	return len(p.NumberCards) > 0
}

// ShowHand displays the player's current hand
func (p *Player) ShowHand() {
	playerIcon := "👤"
	playerLabel := p.Name

	if p.IsComputer() {
		playerIcon = "🤖"
		playerLabel = fmt.Sprintf("%s (%s AI)", p.Name, p.GetAIPersonalityName())
	}

	fmt.Printf("%s %s:\n", playerIcon, playerLabel)

	if len(p.NumberCards) == 0 && len(p.ModifierCards) == 0 {
		fmt.Println("   No cards")
		return
	}

	// Show number cards
	if len(p.NumberCards) > 0 {
		fmt.Print("   Numbers: ")
		for i, card := range p.NumberCards {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(card.String())
		}
		fmt.Println()
	}

	// Show modifier cards
	if len(p.ModifierCards) > 0 {
		fmt.Print("   Modifiers: ")
		for i, card := range p.ModifierCards {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(card.String())
		}
		fmt.Println()
	}

	// Show special status
	if p.HasSecondChance {
		fmt.Println("   🆘 Has Second Chance")
	}

	// Show state
	switch p.State {
	case Stayed:
		fmt.Printf("   ✅ STAYED - Round Score: %d\n", p.RoundScore)
	case Busted:
		fmt.Println("   💥 BUSTED")
	}

	fmt.Println()
}

// GetHandSummary returns a compact summary of the player's hand
func (p *Player) GetHandSummary() string {
	if p.State == Busted {
		return "💥 BUSTED"
	}

	if len(p.NumberCards) == 0 && len(p.ModifierCards) == 0 {
		return "No cards"
	}

	parts := make([]string, 0)

	if len(p.NumberCards) > 0 {
		numbers := make([]string, len(p.NumberCards))
		for i, card := range p.NumberCards {
			numbers[i] = fmt.Sprintf("%d", card.Value)
		}
		parts = append(parts, strings.Join(numbers, ","))
	}

	if len(p.ModifierCards) > 0 {
		mods := make([]string, len(p.ModifierCards))
		for i, card := range p.ModifierCards {
			mods[i] = card.String()
		}
		parts = append(parts, strings.Join(mods, ","))
	}

	result := strings.Join(parts, " | ")

	if p.State == Stayed {
		result += fmt.Sprintf(" (STAYED: %d pts)", p.RoundScore)
	}

	return result
}

// Decision Methods - Handle player input or AI logic internally

// MakeHitStayDecision returns true for hit, false for stay
// Handles both human input and AI decision making internally
func (p *Player) MakeHitStayDecision(gameState *GameState, scanner *bufio.Scanner) (bool, error) {
	p.ShowHand()

	if p.IsComputer() {
		// AI decision making
		shouldHit := p.ShouldHit(gameState)

		fmt.Printf("🤖 %s (%s AI) is thinking", p.Name, p.GetAIPersonalityName())

		// Add some drama with thinking dots
		for i := 0; i < 3; i++ {
			fmt.Print(".")
		}

		if shouldHit {
			fmt.Println(" decides to HIT!")
			return true, nil
		} else {
			fmt.Println(" decides to STAY!")
			return false, nil
		}
	}

	// Human player input
	fmt.Printf("🎯 %s, do you want to (H)it or (S)tay? ", p.Name)
	for {
		if !scanner.Scan() {
			return false, fmt.Errorf("failed to read input")
		}

		choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if choice == "h" || choice == "hit" {
			return true, nil
		}
		if choice == "s" || choice == "stay" {
			return false, nil
		}

		fmt.Print("Please enter 'H' for Hit or 'S' for Stay: ")
	}
}

// ChooseActionTargetInternal selects a target player for action cards
// Handles both human input and AI decision making internally
func (p *Player) ChooseActionTargetInternal(players []*Player, actionType ActionType, gameState *GameState, scanner *bufio.Scanner) (*Player, error) {
	activePlayers := make([]*Player, 0)
	for _, player := range players {
		if player.IsActive() {
			activePlayers = append(activePlayers, player)
		}
	}

	if len(activePlayers) == 0 {
		return nil, fmt.Errorf("no active players")
	}

	if len(activePlayers) == 1 {
		target := activePlayers[0]
		if p.IsComputer() {
			fmt.Printf("   🤖 %s (%s AI) chooses %s\n", p.Name, p.GetAIPersonalityName(), target.Name)
		}
		return target, nil
	}

	if p.IsComputer() {
		// AI target selection
		target := p.ChooseActionTarget(players, actionType, gameState)

		if target != nil {
			fmt.Printf("   🤖 %s (%s AI) chooses %s\n", p.Name, p.GetAIPersonalityName(), target.Name)
			return target, nil
		}

		// Fallback to first active player if AI returns nil
		fmt.Printf("   🤖 %s (%s AI) chooses %s (default)\n", p.Name, p.GetAIPersonalityName(), activePlayers[0].Name)
		return activePlayers[0], nil
	}

	// Human player input
	actionName := map[ActionType]string{
		Freeze:       "Who should be frozen?",
		FlipThree:    "Who should flip three cards?",
		SecondChance: "Who should get the Second Chance card?",
	}

	fmt.Printf("   %s\n", actionName[actionType])
	for i, player := range activePlayers {
		playerType := ""
		if player.IsComputer() {
			playerType = fmt.Sprintf(" (%s AI)", player.GetAIPersonalityName())
		}
		fmt.Printf("   %d) %s%s\n", i+1, player.Name, playerType)
	}

	for {
		fmt.Printf("Enter choice (1-%d): ", len(activePlayers))
		if !scanner.Scan() {
			return nil, fmt.Errorf("failed to read input")
		}

		input := strings.TrimSpace(scanner.Text())
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(activePlayers) {
			fmt.Printf("Please enter a number between 1 and %d: ", len(activePlayers))
			continue
		}

		return activePlayers[choice-1], nil
	}
}

// DecideSecondChanceUsageInternal decides whether to use second chance
// Handles both human input and AI decision making internally
func (p *Player) DecideSecondChanceUsageInternal(duplicateValue int, scanner *bufio.Scanner) bool {
	if p.IsComputer() {
		// AI decision: generally should use Second Chance to avoid busting
		fmt.Printf("   🤖 %s (%s AI) decides to use Second Chance!\n", p.Name, p.GetAIPersonalityName())
		return true
	}

	// Human player input
	fmt.Print("   Use Second Chance? (y/n): ")

	for {
		if !scanner.Scan() {
			return false
		}

		choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if choice == "y" || choice == "yes" {
			return true
		}
		if choice == "n" || choice == "no" {
			return false
		}

		fmt.Print("Please enter 'y' for Yes or 'n' for No: ")
	}
}

// AI Decision Methods - Implement these with your own strategies!

// ShouldHit decides whether the AI player should hit or stay
// Returns true to hit, false to stay
// You can implement different strategies based on p.AIStrategy
func (p *Player) ShouldHit(gameState *GameState) bool {
	if p.IsHuman() {
		return false // This should not be called for human players
	}

	// TODO: Implement your AI logic here based on p.AIStrategy
	// For now, just a simple placeholder:
	currentScore := 0
	for _, card := range p.NumberCards {
		currentScore += card.Value
	}

	// Very basic logic - you should replace this
	switch p.AIStrategy {
	case Conservative:
		return currentScore < 12 // Stay safe
	case Aggressive:
		return currentScore < 20 // Push harder
	case Adaptive:
		return currentScore < 15 // Balanced
	case Chaotic:
		return currentScore < 18 // Unpredictable
	default:
		return currentScore < 15
	}
}

// ChooseActionTarget selects a target player for action cards
// Returns the target player, or nil if no valid target
// You can implement different targeting strategies based on p.AIStrategy
func (p *Player) ChooseActionTarget(players []*Player, actionType ActionType, gameState *GameState) *Player {
	if p.IsHuman() {
		return nil // This should not be called for human players
	}

	activePlayers := make([]*Player, 0)
	for _, player := range players {
		if player.IsActive() {
			activePlayers = append(activePlayers, player)
		}
	}

	if len(activePlayers) == 0 {
		return nil
	}

	if len(activePlayers) == 1 {
		return activePlayers[0]
	}

	// TODO: Implement your AI targeting logic here based on p.AIStrategy and actionType
	// For now, just basic placeholder logic:

	switch actionType {
	case Freeze:
		// Try to freeze the player with the highest current round score
		var target *Player
		maxScore := -1
		for _, player := range activePlayers {
			roundScore := 0
			for _, card := range player.NumberCards {
				roundScore += card.Value
			}
			if roundScore > maxScore {
				maxScore = roundScore
				target = player
			}
		}
		return target

	case FlipThree:
		// Choose based on strategy - you can implement different logic
		switch p.AIStrategy {
		case Conservative:
			return p // Conservative AI might use it on themselves
		case Aggressive:
			// Target player with lowest score to help them catch up (chaos)
			var target *Player
			minTotal := 999
			for _, player := range activePlayers {
				if player.TotalScore < minTotal {
					minTotal = player.TotalScore
					target = player
				}
			}
			return target
		default:
			return activePlayers[0] // Default to first active player
		}

	case SecondChance:
		// Usually give to the player who needs it most
		// You can implement more sophisticated logic
		return activePlayers[0]

	default:
		return activePlayers[0]
	}
}

// GetAIPersonalityName returns a friendly name for the AI personality
func (p *Player) GetAIPersonalityName() string {
	if p.IsHuman() {
		return ""
	}

	switch p.AIStrategy {
	case Conservative:
		return "Cautious"
	case Aggressive:
		return "Risky"
	case Adaptive:
		return "Smart"
	case Chaotic:
		return "Wild"
	default:
		return "Basic"
	}
}

// GameState provides context for AI decision making
type GameState struct {
	Round         int
	Players       []*Player
	ActivePlayers []*Player
	CurrentLeader *Player
	CardsLeft     int
}
