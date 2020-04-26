package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"

	"github.com/google/uuid"

	"github.com/julienschmidt/httprouter"
)

var g *game

type die struct {
	Value     int
	Saved     bool
	Committed bool
	Scored    bool
}

type player struct {
	Name   string
	UUID   string
	Dice   []die
	Score  int
	Scored bool
}

type playerReq struct {
	uuid string
	resp chan *player
}

type game struct {
	Players       []*player
	setPlayerChan chan *player
	getPlayerChan chan playerReq
}

func newPlayer() *player {
	p := &player{}
	for i := 0; i < 6; i++ {
		p.Dice = append(p.Dice, die{})
	}

	return p
}

func gameStateProcessor() {
	g = &game{}
	g.setPlayerChan = make(chan *player, 10)
	g.getPlayerChan = make(chan playerReq, 10)

	for {
		select {
		case preq := <-g.getPlayerChan:
			log.Print("Received get player request")
			for _, player := range g.Players {
				if player.UUID == preq.uuid {
					log.Printf("Responding with player %s, with dice %#v", player.UUID, player.Dice)
					preq.resp <- player
				}
			}
			preq.resp <- nil
		case p := <-g.setPlayerChan:
			log.Print("Received set player request")
			for _, player := range g.Players {
				if player.UUID == p.UUID {
					// update player with state from p
					player.Scored = p.Scored
					for i, die := range p.Dice {
						player.Dice[i].Value = die.Value
						player.Dice[i].Committed = die.Committed
						player.Dice[i].Scored = die.Scored
						player.Dice[i].Saved = die.Saved
					}
					log.Printf("Saved player %#v", player)
					continue
				}
			}
			g.Players = append(g.Players, p)
		}
	}
}

func rollHandler(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Print("Someone requested roll handler")

	// Establish session

	// receive dice state, save to player state

	var p *player

	cookie, err := req.Cookie("UUID")
	if err != nil {
		uuid := uuid.New().String()
		cookie = &http.Cookie{
			Name:   "UUID",
			Value:  uuid,
			Domain: "diceking.online",
		}
		http.SetCookie(w, cookie)
		// get player from gameState
		p = newPlayer()
		p.UUID = uuid
		g.setPlayerChan <- p
	} else {
		pr := playerReq{
			uuid: cookie.Value,
			resp: make(chan *player, 1),
		}
		g.getPlayerChan <- pr
		p = <-pr.resp
	}

	if p == nil {
		log.Print("Player is nil")
		return
	}

	for i, _ := range p.Dice {
		p.Dice[i].Value = int(rand.Int31n(6) + 1)
	}

	g.setPlayerChan <- p

	jsonBytes, err := json.Marshal(p.Dice)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func scoreHandler(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Print("Someone requested score handler")
	w.WriteHeader(http.StatusServiceUnavailable)
}
