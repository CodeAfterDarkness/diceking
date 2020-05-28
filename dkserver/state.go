package main

import (
	"encoding/json"
	"io/ioutil"
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

type playerResponse struct {
	Player   *player
	Messages []string
}

type player struct {
	Name           string
	SessionUUID    string
	Dice           []die
	Score          int
	PotentialScore int
	Scored         bool
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

	// saveTicker := time.NewTicker(time.Second * 10)

	for {
		select {
		case preq := <-g.getPlayerChan:
			log.Print("Received get player request")
			var p *player
			for i, player := range g.Players {
				if player.SessionUUID == preq.uuid {
					//log.Printf("Responding with player %s, with dice %#v", player.SessionUUID, player.Dice)
					preq.resp <- g.Players[i]
					break
				}
			}
			preq.resp <- p
		case p := <-g.setPlayerChan:
			log.Print("Received set player request")
			for pidx, player := range g.Players {
				if player.SessionUUID == p.SessionUUID {
					// update player with state from p
					g.Players[pidx].Scored = p.Scored
					for i, die := range p.Dice {
						g.Players[pidx].Dice[i].Value = die.Value
						g.Players[pidx].Dice[i].Committed = die.Committed
						g.Players[pidx].Dice[i].Scored = die.Scored
						g.Players[pidx].Dice[i].Saved = die.Saved
					}
					//log.Printf("Saved player %#v", g.Players[pidx])
					continue
				}
			}
			g.Players = append(g.Players, p)
			// case <-saveTicker.C:
			// 	f, err := os.OpenFile("gameState.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			// 	if err != nil {
			// 		log.Print(err)
			// 		continue
			// 	}

		}
	}
}

func newSession() *player {
	var p *player
	uuid := uuid.New().String()

	// get player from gameState
	p = newPlayer()
	p.SessionUUID = uuid
	g.setPlayerChan <- p
	return p
}

func rollHandler(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Print("Someone requested roll handler")

	// Establish session

	// receive dice state, save to player state

	p := &player{}

	output := []string{}

	w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, *")

	cookie, err := req.Cookie("UUID")
	if err != nil {
		p = newSession()
		cookie = &http.Cookie{
			Name:   "UUID",
			Value:  p.SessionUUID,
			Domain: req.URL.Hostname(),
		}
		http.SetCookie(w, cookie)
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
		p = newSession()
		cookie = &http.Cookie{
			Name:   "UUID",
			Value:  p.SessionUUID,
			Domain: req.URL.Hostname(),
		}
		http.SetCookie(w, cookie)

		for i, _ := range p.Dice {
			p.Dice[i].Value = int(rand.Int31n(6) + 1)
		}

		//log.Printf("New player session UUID: %s", p.SessionUUID)
	} else {
		//log.Printf("Found existing player %v", p.SessionUUID)

		userDice := []die{}

		jsonBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Print(err)
			return
		}

		err = json.Unmarshal(jsonBytes, &userDice)
		if err != nil {
			log.Print(err)
			return
		}

		var committedDice []die
		var uncommittedDice []die
		for i, d := range p.Dice {
			if d.Committed {
				continue
			}

			if d.Saved {
				log.Printf("User '%s' saved die %d value %d", p.Name, i, d.Value)
				p.Dice[i].Committed = true
				committedDice = append(committedDice, d)
			} else {
				p.Dice[i].Value = int(rand.Int31n(6) + 1)
				uncommittedDice = append(uncommittedDice, d)
			}
		}

		p.PotentialScore += evaluateScore(committedDice, &output)

		score := evaluateScore(uncommittedDice, &output)
		if score == 0 {
			p.PotentialScore = 0
			log.Print("Farkle!")
			p.Scored = true
		}
	}

	//log.Printf("Player rolled dice: %v", p.Dice)

	resp := playerResponse{
		Player:   p,
		Messages: output,
	}

	log.Printf("Player messages: %v", output)
	jsonBytes, err := json.Marshal(resp)
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

	p := &player{}

	output := []string{}

	w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, *")

	cookie, err := req.Cookie("UUID")
	if err != nil {
		p = newSession()
		//log.Printf("Created new session with UUID %s", p.SessionUUID)
		cookie = &http.Cookie{
			Name:   "UUID",
			Value:  p.SessionUUID,
			Domain: req.URL.Hostname(),
		}
		http.SetCookie(w, cookie)
	} else {
		log.Printf("Cookie UUID: %s", cookie.Value)
		pr := playerReq{
			uuid: cookie.Value,
			resp: make(chan *player, 1),
		}
		g.getPlayerChan <- pr
		//log.Print("Waiting for player response")
		p = <-pr.resp
		//log.Print("Got player response")
	}

	if p == nil {
		log.Print("Player is nil")
		p = newSession()
		cookie = &http.Cookie{
			Name:   "UUID",
			Value:  p.SessionUUID,
			Domain: req.URL.Hostname(),
		}
		http.SetCookie(w, cookie)
	} else {
		//log.Printf("Found existing player %v", p.SessionUUID)
	}

	userDice := []die{}

	jsonBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Print(err)
		return
	}

	err = json.Unmarshal(jsonBytes, &userDice)
	if err != nil {
		log.Print(err)
		return
	}

	var committedDice []die

	for _, d := range p.Dice {
		if d.Committed {
			committedDice = append(committedDice, d)
		}
	}

	score := evaluateScore(committedDice, &output)
	if score == 0 {
		p.PotentialScore = 0
	} else {
		p.Score += p.PotentialScore + score
		p.PotentialScore = 0
		log.Printf("User '%s' scored %d, score is now %d", p.Name, score, p.Score)
	}

	for i, _ := range p.Dice {
		p.Dice[i].Committed = false
		p.Dice[i].Saved = false
		p.Dice[i].Scored = false
		p.Dice[i].Value = int(rand.Int31n(6) + 1)
	}

	g.setPlayerChan <- p

	resp := playerResponse{
		Player:   p,
		Messages: output,
	}
	jsonBytes, err = json.Marshal(resp)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)

}
