package main

import (
	"math/rand"
	"time"
)

type die struct {
	Value     int
	Saved     bool
	Committed bool
	Scored    bool
}

type player struct {
	Dice   []die
	Score  int
	Scored bool
}

type game struct {
	Players []player
}

func main() {
	rand.Seed(time.Now().UnixNano())
}

func rollDie() int {
	return int(rand.Int31n(5) + 1)
}
