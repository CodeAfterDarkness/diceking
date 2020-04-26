package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

type die struct {
	Value     int
	Saved     bool
	Committed bool
	Scored    bool
}

type player struct {
	Name        string
	SessionGUID string
	Dice        []die
	Score       int
	Scored      bool
}

func newPlayer() player {
	p := player{}
	for i := 0; i < 6; i++ {
		p.Dice = append(p.Dice, die{})
	}

	return p
}

type game struct {
	Players []player
}

var router *httprouter.Router

func main() {
	rand.Seed(time.Now().UnixNano())

	router = httprouter.New()

	router.GET("/gracefullyRestart", restartHandler)

	router.GET("/roll", rollHandler)
	router.GET("/score", scoreHandler)

	router.GET("/css/*file", cssHandler)

	router.GET("/", baseHandler)

	addr := ":80"
	ln, err := createOrImportListener(addr)
	if err != nil {
		log.Print(err)
	}

	server := startServer(addr, ln)
	err = waitForSignals(addr, ln, server)
	if err != nil {
		fmt.Printf("Exiting: %v\n", err)
		return
	}
	fmt.Printf("Exiting.\n")
}

func cssHandler(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Print("Someone requested css")

	log.Printf("Req: %s", req.URL.String())

	fileBytes, err := ioutil.ReadFile("css" + params.ByName("file"))
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "text/css")
	w.Write(fileBytes)
}

func baseHandler(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Print("Someone requested base")

	log.Printf("Req: %s", req.URL.String())

	fileBytes, err := ioutil.ReadFile("html/index.html")
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(fileBytes)

}

func sessionHandler(session string) (player, error) {
	players := []player{
		{},
	}

	player := players[0]

	return player, nil
}

func rollHandler(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Print("Someone requested roll handler")

	// Establish session

	// receive dice state, save to player state

	p := newPlayer()

	for i, _ := range p.Dice {
		p.Dice[i].Value = int(rand.Int31n(5) + 1)
	}

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
