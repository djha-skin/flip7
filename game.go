package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Game represents the main game state
type Game struct {
	players   []*Player
	deck      *Deck
	round     int
	dealerIdx int
	scanner   *bufio.Scanner
}

// NewGame creates a new Flip 7 game instance
func NewGame() *Game {
	return &Game{
		players: make([]*Player, 0),
		deck:    NewDeck(),
		round:   1,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// Run starts the main game loop
func (g *Game) Run() error {
	fmt.Println("How many players? (3-8): ")
	numPlayers, err := g.getIntInput(3, 8)
	if err != nil {
		return err
	}

	// Initialize players
	for i := 0; i < numPlayers; i++ {
		fmt.Printf("Enter name for Player %d: ", i+1)
		name, err := g.getStringInput()
		if err != nil {
			return err
		}
		g.players = append(g.players, NewPlayer(name))
	}

	fmt.Println("\n🎮 Starting Flip 7! First to 200 points wins!")

	// Main game loop
	for !g.hasWinner() {
		fmt.Printf("\n" + strings.Repeat("=", 50))
		fmt.Printf("\n🎯 ROUND %d\n", g.round)
		fmt.Printf(strings.Repeat("=", 50) + "\n")

		if err := g.playRound(); err != nil {
			return err
		}

		g.showScores()
		g.nextRound()
	}

	winner := g.getWinner()
	fmt.Printf("\n🎉 GAME OVER! %s wins with %d points! 🎉\n", winner.Name, winner.TotalScore)

	return nil
}

// Helper methods for input handling
func (g *Game) getIntInput(min, max int) (int, error) {
	for {
		if !g.scanner.Scan() {
			return 0, fmt.Errorf("failed to read input")
		}

		input := strings.TrimSpace(g.scanner.Text())
		num, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Please enter a valid number between %d and %d: ", min, max)
			continue
		}

		if num < min || num > max {
			fmt.Printf("Please enter a number between %d and %d: ", min, max)
			continue
		}

		return num, nil
	}
}

func (g *Game) getStringInput() (string, error) {
	if !g.scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}
	return strings.TrimSpace(g.scanner.Text()), nil
}

func (g *Game) hasWinner() bool {
	for _, player := range g.players {
		if player.TotalScore >= 200 {
			return true
		}
	}
	return false
}

func (g *Game) getWinner() *Player {
	var winner *Player
	maxScore := -1

	for _, player := range g.players {
		if player.TotalScore > maxScore {
			maxScore = player.TotalScore
			winner = player
		}
	}

	return winner
}

func (g *Game) showScores() {
	fmt.Println("\n📊 Current Scores:")
	fmt.Println(strings.Repeat("-", 30))
	for _, player := range g.players {
		fmt.Printf("%-15s: %3d points\n", player.Name, player.TotalScore)
	}
	fmt.Println(strings.Repeat("-", 30))
}

func (g *Game) nextRound() {
	g.round++
	g.dealerIdx = (g.dealerIdx + 1) % len(g.players)

	// Reset players for new round
	for _, player := range g.players {
		player.ResetForNewRound()
	}

	// Check if deck needs reshuffling
	if g.deck.CardsLeft() < len(g.players)*5 {
		fmt.Println("🔀 Reshuffling deck...")
		g.deck.Reshuffle()
	}
}

func (g *Game) playRound() error {
	fmt.Printf("Dealer: %s\n\n", g.players[g.dealerIdx].Name)

	// Deal initial cards
	if err := g.dealInitialCards(); err != nil {
		return err
	}

	// Play turns until round ends
	if err := g.playTurns(); err != nil {
		return err
	}

	// Calculate scores
	g.calculateRoundScores()

	return nil
}

func (g *Game) dealInitialCards() error {
	fmt.Println("🃏 Dealing initial cards...")

	// Deal one card to each player
	for i := 0; i < len(g.players); i++ {
		playerIdx := (g.dealerIdx + 1 + i) % len(g.players)
		player := g.players[playerIdx]

		card := g.deck.DrawCard()
		if card == nil {
			return fmt.Errorf("deck is empty")
		}

		fmt.Printf("   %s draws %s\n", player.Name, card.String())

		// Handle action cards immediately
		if card.IsActionCard() {
			if err := g.handleActionCard(player, card); err != nil {
				return err
			}
		} else {
			if err := player.AddCard(card); err != nil {
				return g.handleCardAddError(player, card, err)
			}
		}
	}

	fmt.Println()
	g.showAllHands()
	return nil
}

func (g *Game) playTurns() error {
	for g.hasActivePlayers() {
		for i := 0; i < len(g.players); i++ {
			playerIdx := (g.dealerIdx + 1 + i) % len(g.players)
			player := g.players[playerIdx]

			if !player.IsActive() {
				continue
			}

			// Player must hit if they have no number cards
			if !player.HasCards() {
				fmt.Printf("🎯 %s has no number cards and must HIT\n", player.Name)
				if err := g.playerHit(player); err != nil {
					return err
				}
				continue
			}

			// Ask player to hit or stay
			choice, err := g.getPlayerChoice(player)
			if err != nil {
				return err
			}

			if choice == "h" {
				if err := g.playerHit(player); err != nil {
					return err
				}
			} else {
				g.playerStay(player)
			}

			if !g.hasActivePlayers() {
				break
			}
		}
	}

	return nil
}

func (g *Game) calculateRoundScores() {
	fmt.Println("📊 Calculating round scores...")
	fmt.Println(strings.Repeat("-", 40))

	for _, player := range g.players {
		roundScore := player.CalculateRoundScore()
		player.AddToTotalScore()

		fmt.Printf("%s: %d points this round (Total: %d)\n",
			player.Name, roundScore, player.TotalScore)
	}
	fmt.Println(strings.Repeat("-", 40))
}

// Helper methods for gameplay

func (g *Game) hasActivePlayers() bool {
	for _, player := range g.players {
		if player.IsActive() {
			return true
		}
	}
	return false
}

func (g *Game) showAllHands() {
	for _, player := range g.players {
		player.ShowHand()
	}
}

func (g *Game) getPlayerChoice(player *Player) (string, error) {
	player.ShowHand()
	fmt.Printf("🎯 %s, do you want to (H)it or (S)tay? ", player.Name)

	for {
		choice, err := g.getStringInput()
		if err != nil {
			return "", err
		}

		choice = strings.ToLower(strings.TrimSpace(choice))
		if choice == "h" || choice == "hit" {
			return "h", nil
		}
		if choice == "s" || choice == "stay" {
			return "s", nil
		}

		fmt.Print("Please enter 'H' for Hit or 'S' for Stay: ")
	}
}

func (g *Game) playerHit(player *Player) error {
	card := g.deck.DrawCard()
	if card == nil {
		return fmt.Errorf("deck is empty")
	}

	fmt.Printf("   %s draws %s\n", player.Name, card.String())

	if card.IsActionCard() {
		return g.handleActionCard(player, card)
	}

	if err := player.AddCard(card); err != nil {
		return g.handleCardAddError(player, card, err)
	}

	return nil
}

func (g *Game) playerStay(player *Player) {
	player.Stay()
	player.CalculateRoundScore()
	fmt.Printf("   %s stays with %d points\n", player.Name, player.RoundScore)
}

func (g *Game) handleActionCard(player *Player, card *Card) error {
	fmt.Printf("   🎲 Action card! %s\n", card.String())

	switch card.Action {
	case Freeze:
		return g.handleFreezeCard(player, card)
	case FlipThree:
		return g.handleFlipThreeCard(player, card)
	case SecondChance:
		return g.handleSecondChanceCard(player, card)
	}

	return nil
}

func (g *Game) handleFreezeCard(player *Player, card *Card) error {
	target, err := g.chooseActionTarget(player, "Who should be frozen?")
	if err != nil {
		return err
	}

	target.Stay()
	target.CalculateRoundScore()
	fmt.Printf("   ❄️ %s is frozen and stays with %d points!\n", target.Name, target.RoundScore)

	g.deck.DiscardCard(card)
	return nil
}

func (g *Game) handleFlipThreeCard(player *Player, card *Card) error {
	target, err := g.chooseActionTarget(player, "Who should flip three cards?")
	if err != nil {
		return err
	}

	fmt.Printf("   🎲 %s must flip 3 cards!\n", target.Name)

	for i := 0; i < 3; i++ {
		if !target.IsActive() {
			break
		}

		drawnCard := g.deck.DrawCard()
		if drawnCard == nil {
			break
		}

		fmt.Printf("      Card %d: %s\n", i+1, drawnCard.String())

		if drawnCard.IsActionCard() {
			// Handle nested action cards after all 3 cards are drawn
			if err := g.handleActionCard(target, drawnCard); err != nil {
				if strings.Contains(err.Error(), "flip7") {
					fmt.Printf("   🎉 %s achieved FLIP 7!\n", target.Name)
					return err
				}
				return err
			}
		} else {
			if err := target.AddCard(drawnCard); err != nil {
				if err := g.handleCardAddError(target, drawnCard, err); err != nil {
					return err
				}
				break
			}
		}
	}

	g.deck.DiscardCard(card)
	return nil
}

func (g *Game) handleSecondChanceCard(player *Player, card *Card) error {
	target, err := g.chooseActionTarget(player, "Who should get the Second Chance card?")
	if err != nil {
		return err
	}

	if err := target.AddCard(card); err != nil {
		if strings.Contains(err.Error(), "second_chance_duplicate") {
			// Find another player who doesn't have second chance
			for _, p := range g.players {
				if p.IsActive() && !p.HasSecondChance && p != target {
					fmt.Printf("   🆘 %s already has Second Chance, giving to %s instead\n", target.Name, p.Name)
					return p.AddCard(card)
				}
			}
			// No one can take it, discard
			fmt.Printf("   🆘 No one can take the Second Chance card, discarding\n")
			g.deck.DiscardCard(card)
		}
		return err
	}

	fmt.Printf("   🆘 %s receives a Second Chance card!\n", target.Name)
	return nil
}

func (g *Game) chooseActionTarget(player *Player, prompt string) (*Player, error) {
	activePlayers := make([]*Player, 0)
	for _, p := range g.players {
		if p.IsActive() {
			activePlayers = append(activePlayers, p)
		}
	}

	if len(activePlayers) == 0 {
		return nil, fmt.Errorf("no active players")
	}

	if len(activePlayers) == 1 {
		return activePlayers[0], nil
	}

	fmt.Printf("   %s\n", prompt)
	for i, p := range activePlayers {
		fmt.Printf("   %d) %s\n", i+1, p.Name)
	}

	choice, err := g.getIntInput(1, len(activePlayers))
	if err != nil {
		return nil, err
	}

	return activePlayers[choice-1], nil
}

func (g *Game) handleCardAddError(player *Player, card *Card, err error) error {
	if strings.Contains(err.Error(), "flip7") {
		fmt.Printf("   🎉 %s achieved FLIP 7 and wins the round!\n", player.Name)
		return err
	}

	if strings.Contains(err.Error(), "duplicate_with_second_chance") {
		parts := strings.Split(err.Error(), ":")
		if len(parts) > 1 {
			duplicateValue := parts[1]
			fmt.Printf("   💥 %s drew a duplicate %s but has Second Chance!\n", player.Name, duplicateValue)
			fmt.Print("   Use Second Chance? (y/n): ")

			choice, err := g.getStringInput()
			if err != nil {
				return err
			}

			if strings.ToLower(strings.TrimSpace(choice)) == "y" {
				if value, parseErr := strconv.Atoi(duplicateValue); parseErr == nil {
					player.UseSecondChance(value)
					fmt.Printf("   🆘 %s used Second Chance to avoid busting!\n", player.Name)
					g.deck.DiscardCard(card) // Discard the duplicate
					return nil
				}
			}
		}

		// If not using second chance, player busts
		player.State = Busted
		fmt.Printf("   💥 %s chose not to use Second Chance and busts!\n", player.Name)
		return nil
	}

	if strings.Contains(err.Error(), "bust") {
		fmt.Printf("   💥 %s busts and is out of the round!\n", player.Name)
		return nil
	}

	return err
}
