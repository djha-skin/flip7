package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("🎴 Welcome to Flip 7!")
	fmt.Println("Press your luck and flip your way to 200 points!")
	fmt.Println()

	game := NewGame()
	if err := game.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
