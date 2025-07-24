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
	debugMode bool
}

// NewGame creates a new Flip 7 game instance
func NewGame() *Game {
	return &Game{
		players:   make([]*Player, 0),
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
	fmt.Println(strings.Repeat("-", 40))
	for _, player := range g.players {
		icon := "👤"
		if player.PlayerType == Computer {
			icon = "🤖"
		}
		fmt.Printf("%s %-20s: %3d points\n", icon, player.Name, player.TotalScore)
	}
	fmt.Println(strings.Repeat("-", 40))
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
	gameState := g.buildGameState()
	shouldHit, err := player.MakeHitStayDecision(gameState, g.scanner)
	if err != nil {
		return "", err
	}

	if shouldHit {
		return "h", nil
	} else {
		return "s", nil
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
	target, err := g.chooseActionTarget(player, "Who should be frozen?", Freeze)
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
	target, err := g.chooseActionTarget(player, "Who should flip three cards?", FlipThree)
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
					g.endRoundForFlip7(target)
					break // End the Flip Three loop
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
	// Try to give it to the player who drew it first
	if !player.HasSecondChance {
		if err := player.AddCard(card); err != nil {
			return err
		}
		fmt.Printf("   🆘 %s receives a Second Chance card!\n", player.Name)
		return nil
	}

	// Player already has second chance, need to give it to someone else
	fmt.Printf("   🆘 %s already has Second Chance, must give to another player\n", player.Name)

	target, err := g.chooseActionTarget(player, "Who should get the Second Chance card?", SecondChance)
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

func (g *Game) chooseActionTarget(player *Player, prompt string, actionType ActionType) (*Player, error) {
	gameState := g.buildGameState()
	return player.ChooseActionTargetInternal(g.players, actionType, gameState, g.scanner)
}

func (g *Game) handleCardAddError(player *Player, card *Card, err error) error {
	if strings.Contains(err.Error(), "flip7") {
		fmt.Printf("   🎉 %s achieved FLIP 7 and wins the round!\n", player.Name)
		// Mark all other players as non-active to end the round
		g.endRoundForFlip7(player)
		return nil // Don't propagate the error, just end the round
	}

	if strings.Contains(err.Error(), "duplicate_with_second_chance") {
		parts := strings.Split(err.Error(), ":")
		if len(parts) > 1 {
			duplicateValue := parts[1]
			fmt.Printf("   💥 %s drew a duplicate %s but has Second Chance!\n", player.Name, duplicateValue)

			if value, parseErr := strconv.Atoi(duplicateValue); parseErr == nil {
				useSecondChance := player.DecideSecondChanceUsageInternal(value, g.scanner)

				if useSecondChance {
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
		strategy, name, err := g.getComputerPlayerSetup(i + 1)
		if err != nil {
			return err
		}
		g.players = append(g.players, NewComputerPlayer(name, strategy))
		fmt.Printf("  → Added: %s (%s AI)\n", name, g.players[len(g.players)-1].GetAIPersonalityName())
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
func (g *Game) getComputerPlayerSetup(computerNum int) (AIStrategy, string, error) {
	fmt.Printf("\nComputer Player %d:\n", computerNum)
	fmt.Println("Choose AI strategy:")
	fmt.Println("  1) Conservative (plays it safe)")
	fmt.Println("  2) Aggressive (takes big risks)")
	fmt.Println("  3) Adaptive (adjusts to game state)")
	fmt.Println("  4) Chaotic (unpredictable)")
	fmt.Print("Enter choice (1-4): ")

	choice, err := g.getIntInput(1, 4)
	if err != nil {
		return Conservative, "", err
	}

	strategy := AIStrategy(choice - 1)

	// Generate a name or let user customize
	fmt.Print("Use default name? (y/n): ")
	useDefault, err := g.getStringInput()
	if err != nil {
		return strategy, "", err
	}

	var name string
	if strings.ToLower(strings.TrimSpace(useDefault)) == "y" {
		name = g.generateComputerName(strategy, computerNum)
	} else {
		fmt.Print("Enter custom name: ")
		name, err = g.getStringInput()
		if err != nil {
			return strategy, "", err
		}
	}

	return strategy, name, nil
}

// generateComputerName creates a default name for computer players
func (g *Game) generateComputerName(strategy AIStrategy, num int) string {
	baseNames := map[AIStrategy][]string{
		Conservative: {"Cautious Carl", "Safe Sally", "Prudent Pete", "Careful Claire"},
		Aggressive:   {"Risky Rita", "Bold Bob", "Daring Dave", "Fearless Fiona"},
		Adaptive:     {"Smart Sam", "Clever Cathy", "Tactical Tom", "Strategic Sue"},
		Chaotic:      {"Wild Will", "Crazy Kate", "Random Rick", "Chaotic Charlie"},
	}

	names := baseNames[strategy]
	if num-1 < len(names) {
		return names[num-1]
	}

	// Fallback for additional computer players
	prefix := map[AIStrategy]string{
		Conservative: "Cautious",
		Aggressive:   "Risky",
		Adaptive:     "Smart",
		Chaotic:      "Wild",
	}[strategy]

	return fmt.Sprintf("%s Bot %d", prefix, num)
}

// buildGameState creates a GameState for AI decision making
func (g *Game) buildGameState() *GameState {
	activePlayers := make([]*Player, 0)
	for _, p := range g.players {
		if p.IsActive() {
			activePlayers = append(activePlayers, p)
		}
	}

	var currentLeader *Player
	maxScore := -1
	for _, p := range g.players {
		if p.TotalScore > maxScore {
			maxScore = p.TotalScore
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
func (g *Game) endRoundForFlip7(flip7Player *Player) {
	for _, player := range g.players {
		if player != flip7Player && player.IsActive() {
			player.Stay()
			player.CalculateRoundScore()
		}
	}
}
