package main

import (
	"math/rand"
	"time"
)

// Deck represents the game deck
type Deck struct {
	cards    []*Card
	discards []*Card
	rng      *rand.Rand
}

// NewDeck creates a new deck with the correct card distribution for Flip 7
func NewDeck() *Deck {
	deck := &Deck{
		cards:    make([]*Card, 0),
		discards: make([]*Card, 0),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	deck.createCards()
	deck.Shuffle()

	return deck
}

// createCards creates all cards with the correct distributions
func (d *Deck) createCards() {
	// Number cards: each number has as many cards as its value
	// 12 has 12 copies, 11 has 11 copies, etc., down to 0 which has 1 copy
	for value := 0; value <= 12; value++ {
		count := value
		if value == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			d.cards = append(d.cards, NewNumberCard(value))
		}
	}

	// Score Modifier Cards (6 total)
	d.cards = append(d.cards, NewModifierCard(Plus2))
	d.cards = append(d.cards, NewModifierCard(Plus4))
	d.cards = append(d.cards, NewModifierCard(Plus6))
	d.cards = append(d.cards, NewModifierCard(Plus8))
	d.cards = append(d.cards, NewModifierCard(Plus10))
	d.cards = append(d.cards, NewModifierCard(Multiply2))

	// Action Cards (3 of each type = 9 total)
	for i := 0; i < 3; i++ {
		d.cards = append(d.cards, NewActionCard(Freeze))
		d.cards = append(d.cards, NewActionCard(FlipThree))
		d.cards = append(d.cards, NewActionCard(SecondChance))
	}
}

// Shuffle shuffles the deck
func (d *Deck) Shuffle() {
	d.rng.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

// DrawCard draws the top card from the deck
func (d *Deck) DrawCard() *Card {
	if len(d.cards) == 0 {
		return nil
	}

	card := d.cards[len(d.cards)-1]
	d.cards = d.cards[:len(d.cards)-1]
	return card
}

// DiscardCard adds a card to the discard pile
func (d *Deck) DiscardCard(card *Card) {
	if card != nil {
		d.discards = append(d.discards, card)
	}
}

// Reshuffle reshuffles the discard pile back into the deck
func (d *Deck) Reshuffle() {
	d.cards = append(d.cards, d.discards...)
	d.discards = make([]*Card, 0)
	d.Shuffle()
}

// CardsLeft returns the number of cards remaining in the deck
func (d *Deck) CardsLeft() int {
	return len(d.cards)
}

// TotalCards returns the total number of cards (deck + discards)
func (d *Deck) TotalCards() int {
	return len(d.cards) + len(d.discards)
}
