package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	rand.Seed(time.Now().UnixNano())

	go gameStateProcessor()

	router = httprouter.New()

	router.GET("/gracefullyRestart", restartHandler)
	router.POST("/roll", rollHandler)
	router.POST("/score", scoreHandler)
	router.GET("/css/*file", cssHandler)
	router.GET("/js/*file", jsHandler)
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

var router *httprouter.Router

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

func jsHandler(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Print("Someone requested js")

	log.Printf("Req: %s", req.URL.String())

	fileBytes, err := ioutil.ReadFile("js" + params.ByName("file"))
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/js")
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
