package main

import (
	"fmt"
	"strings"
)

// PlayerState represents the current state of a player in a round
type PlayerState int

const (
	Active PlayerState = iota
	Stayed
	Busted
)

// Player represents a game player
type Player struct {
	Name            string
	TotalScore      int
	RoundScore      int
	NumberCards     []*Card
	ModifierCards   []*Card
	ActionCards     []*Card
	State           PlayerState
	HasSecondChance bool
}

// NewPlayer creates a new player
func NewPlayer(name string) *Player {
	return &Player{
		Name:            name,
		TotalScore:      0,
		RoundScore:      0,
		NumberCards:     make([]*Card, 0),
		ModifierCards:   make([]*Card, 0),
		ActionCards:     make([]*Card, 0),
		State:           Active,
		HasSecondChance: false,
	}
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
	fmt.Printf("👤 %s:\n", p.Name)

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
