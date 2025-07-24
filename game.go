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
	players   []PlayerInterface
	deck      *Deck
	round     int
	dealerIdx int
	scanner   *bufio.Scanner
	debugMode bool
}

// NewGame creates a new Flip 7 game instance
func NewGame() *Game {
	return &Game{
		players:   make([]PlayerInterface, 0),
		deck:      NewDeck(),
		round:     1,
		scanner:   bufio.NewScanner(os.Stdin),
		debugMode: false,
	}
}

// SetDebugMode enables or disables debug mode
func (g *Game) SetDebugMode(debug bool) {
	g.debugMode = debug
	g.deck.SetDebugMode(debug, g.scanner)
}

// Run starts the main game loop
func (g *Game) Run() error {
	// Setup players
	if err := g.setupPlayers(); err != nil {
		return err
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
	fmt.Printf("\n🎉 GAME OVER! %s wins with %d points! 🎉\n", winner.GetName(), winner.GetTotalScore())

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
		if player.GetTotalScore() >= 200 {
			return true
		}
	}
	return false
}

func (g *Game) getWinner() PlayerInterface {
	var winner PlayerInterface
	maxScore := -1

	for _, player := range g.players {
		if player.GetTotalScore() > maxScore {
			maxScore = player.GetTotalScore()
			winner = player
		}
	}

	return winner
}

func (g *Game) showScores() {
	fmt.Println("\n📊 Current Scores:")
	fmt.Println(strings.Repeat("-", 40))
	for _, player := range g.players {
		icon := player.GetPlayerIcon()
		fmt.Printf("%s %-20s: %3d points\n", icon, player.GetName(), player.GetTotalScore())
	}
	fmt.Println(strings.Repeat("-", 40))
}

func (g *Game) nextRound() {
	g.round++
	g.dealerIdx = (g.dealerIdx + 1) % len(g.players)

	// Reset players for new round
	for _, player := range g.players {
		discardedCards := player.ResetForNewRound()
		for _, card := range discardedCards {
			g.deck.DiscardCard(card)
		}
	}

	totalCards := g.deck.CardsLeft() + len(g.deck.discards)
	for _, player := range g.players {
		totalCards += len(player.GetHand())
	}

	if totalCards != g.deck.OriginalTotal {
		totals := map[string]int{}
		for _, card := range g.deck.cards {
			totals[card.String()]++
		}
		for _, card := range g.deck.discards {
			totals[card.String()]++
		}
		for _, player := range g.players {
			for _, card := range player.GetHand() {
				totals[card.String()]++
			}
		}
		fmt.Println(totals)
		panic(fmt.Sprintf("Total cards is not the original total. Cards are disappearing! found: %d != excpected: %d", totalCards, g.deck.OriginalTotal))
	}
}

func (g *Game) playRound() error {
	fmt.Printf("Dealer: %s\n\n", g.players[g.dealerIdx].GetName())

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

		fmt.Printf("   %s draws %s\n", player.GetName(), card.String())

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
				fmt.Printf("🎯 %s has no number cards and must HIT\n", player.GetName())
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
			player.GetName(), roundScore, player.GetTotalScore())
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

func (g *Game) getPlayerChoice(player PlayerInterface) (string, error) {
	gameState := g.buildGameState()
	shouldHit, err := player.MakeHitStayDecision(gameState)
	if err != nil {
		return "", err
	}

	if shouldHit {
		return "h", nil
	} else {
		return "s", nil
	}
}

func (g *Game) playerHit(player PlayerInterface) error {
	card := g.deck.DrawCard()
	if card == nil {
		return fmt.Errorf("deck is empty")
	}

	fmt.Printf("   %s draws %s\n", player.GetName(), card.String())

	if card.IsActionCard() {
		return g.handleActionCard(player, card)
	}

	if err := player.AddCard(card); err != nil {
		return g.handleCardAddError(player, card, err)
	}

	return nil
}

func (g *Game) playerStay(player PlayerInterface) {
	player.Stay()
	player.CalculateRoundScore()
	fmt.Printf("   %s stays with %d points\n", player.GetName(), player.CalculateRoundScore())
}

func (g *Game) handleActionCard(player PlayerInterface, card *Card) error {
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

func (g *Game) handleFreezeCard(player PlayerInterface, card *Card) error {
	target, err := g.chooseActionTarget(player, "Who should be frozen?", Freeze)
	if err != nil {
		g.deck.DiscardCard(card) // Discard card even if target selection fails
		return err
	}

	target.Stay()
	target.CalculateRoundScore()
	fmt.Printf("   ❄️ %s is frozen and stays with %d points!\n", target.GetName(), target.CalculateRoundScore())

	g.deck.DiscardCard(card)
	return nil
}

func (g *Game) handleFlipThreeCard(player PlayerInterface, card *Card) error {
	target, err := g.chooseActionTarget(player, "Who should flip three cards?", FlipThree)
	if err != nil {
		g.deck.DiscardCard(card) // Discard card even if target selection fails
		return err
	}

	fmt.Printf("   🎲 %s must flip 3 cards!\n", target.GetName())

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
					fmt.Printf("   🎉 %s achieved FLIP 7!\n", target.GetName())
					g.endRoundForFlip7(target)
					break // End the Flip Three loop
				}
				// Discard the action card if there was an error handling it
				g.deck.DiscardCard(drawnCard)
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

func (g *Game) handleSecondChanceCard(player PlayerInterface, card *Card) error {
	// Try to give it to the player who drew it first
	if !player.HasSecondChance() {
		fmt.Printf("   🆘 %s receives a Second Chance card!\n", player.GetName())
		if err := player.AddCard(card); err != nil {
			g.deck.DiscardCard(card)
			return err
		}
		return nil
	}

	// Player already has second chance, need to give it to someone else
	fmt.Printf("   🆘 %s already has Second Chance, must give to another player\n", player.GetName())

	target, err := player.ChoosePositiveActionTarget(g.buildGameState(), SecondChance)
	if err != nil {
		fmt.Printf("   🆘 No one can take the Second Chance card, discarding\n")
		g.deck.DiscardCard(card)
		return err
	}

	if err := target.AddCard(card); err != nil {
		g.deck.DiscardCard(card)
		fmt.Printf("   🆘 %s cannot take the Second Chance card, discarding\n", player.GetName())
		return err
	}

	fmt.Printf("   🆘 %s receives a Second Chance card!\n", target.GetName())
	return nil
}

func (g *Game) chooseActionTarget(player PlayerInterface, prompt string, actionType ActionType) (PlayerInterface, error) {
	gameState := g.buildGameState()
	return player.ChooseActionTarget(gameState, actionType)
}

func (g *Game) handleCardAddError(player PlayerInterface, card *Card, err error) error {
	if strings.Contains(err.Error(), "flip7") {
		fmt.Printf("   🎉 %s achieved FLIP 7 and wins the round!\n", player.GetName())
		// Mark all other players as non-active to end the round
		g.endRoundForFlip7(player)
		return nil // Don't propagate the error, just end the round
	}

	if strings.Contains(err.Error(), "duplicate_with_second_chance") {
		fmt.Printf("   💥 %s drew a duplicate %s but has Second Chance!\n", player.GetName(), card)
		secondChanceCard := player.UseSecondChance()
		g.deck.DiscardCard(secondChanceCard) // Discard the second chance card
		g.deck.DiscardCard(card)             // Discard the duplicate
		return nil
	}

	if strings.Contains(err.Error(), "bust") {
		g.deck.DiscardCard(card) // Discard the duplicate
		fmt.Printf("   💥 %s busts and is out of the round!\n", player.GetName())
		return nil
	}

	return err
}

// setupPlayers handles the initial player setup (human vs computer)
func (g *Game) setupPlayers() error {
	fmt.Println("How many players total? (2-8): ")
	numPlayers, err := g.getIntInput(2, 8)
	if err != nil {
		return err
	}

	fmt.Printf("How many human players? (0-%d): ", numPlayers)
	numHumans, err := g.getIntInput(0, numPlayers)
	if err != nil {
		return err
	}

	numComputers := numPlayers - numHumans

	// Setup human players
	for i := 0; i < numHumans; i++ {
		fmt.Printf("Enter name for Human Player %d: ", i+1)
		name, err := g.getStringInput()
		if err != nil {
			return err
		}
		g.players = append(g.players, NewHumanPlayer(name, g.scanner))
	}

	// Setup computer players
	for i := 0; i < numComputers; i++ {
		name, strategy, actionTargetStrategy, positiveActionTargetStrategy, err := g.getComputerPlayerSetup(i + 1)
		if err != nil {
			return err
		}
		g.players = append(g.players, NewComputerPlayer(name, strategy, actionTargetStrategy, positiveActionTargetStrategy))
		fmt.Printf("  → Added: %s (%s AI)\n", name, g.players[len(g.players)-1].GetName())
	}

	if numHumans == 0 {
		fmt.Printf("\n🎮 Starting AI-only Flip 7 with %d computer players!\n", numComputers)
		fmt.Println("🍿 Sit back and watch the AIs battle it out!")
	} else {
		fmt.Printf("\n🎮 Starting Flip 7 with %d humans and %d computers!\n", numHumans, numComputers)
	}
	return nil
}

// getComputerPlayerSetup handles setup for a single computer player
func (g *Game) getComputerPlayerSetup(computerNum int) (string, HitOrStayStrategy, ActionTargetStrategy, ActionTargetStrategy, error) {
	fmt.Printf("\nComputer Player %d:\n", computerNum)
	fmt.Println("Choose AI strategy:")
	fmt.Println("  1) Plays to 20")
	fmt.Println("  2) Plays to 25")
	fmt.Println("  3) Plays to 30")
	fmt.Println("  4) FLIP 7")
	fmt.Println("  5) Random")
	fmt.Print("Enter choice (1-5): ")

	choice, err := g.getIntInput(1, 5)
	if err != nil {
		choice = 5
	}

	switch choice {
	case 1:
		return "Plays to 20", PlayRoundTo(20), TargetLeaderStrategy, TargetLastPlaceStrategy, nil
	case 2:
		return "Plays to 25", PlayRoundTo(25), TargetLeaderStrategy, TargetLastPlaceStrategy, nil
	case 3:
		return "Plays to 30", PlayRoundTo(30), TargetLeaderStrategy, TargetLastPlaceStrategy, nil
	case 4:
		return "FLIP 7", AlwaysHitStrategy, TargetLeaderStrategy, TargetLastPlaceStrategy, nil
	case 5:
		fallthrough
	default:
		return "Random", RandomHitOrStayStrategy, TargetRandomStrategy, TargetRandomStrategy, nil
	}
}

// buildGameState creates a GameState for AI decision making
func (g *Game) buildGameState() *GameState {
	activePlayers := make([]PlayerInterface, 0)
	for _, p := range g.players {
		if p.IsActive() {
			activePlayers = append(activePlayers, p)
		}
	}

	var currentLeader PlayerInterface
	maxScore := -1
	for _, p := range g.players {
		if p.GetTotalScore() > maxScore {
			maxScore = p.GetTotalScore()
			currentLeader = p
		}
	}

	return &GameState{
		Round:         g.round,
		Players:       g.players,
		ActivePlayers: activePlayers,
		CurrentLeader: currentLeader,
		CardsLeft:     g.deck.CardsLeft(),
	}
}

// endRoundForFlip7 marks all players except the Flip 7 achiever as non-active
func (g *Game) endRoundForFlip7(flip7Player PlayerInterface) {
	for _, player := range g.players {
		if player != flip7Player && player.IsActive() {
			player.Stay()
			player.CalculateRoundScore()
		}
	}
}
