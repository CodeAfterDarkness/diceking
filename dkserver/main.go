package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
	uuid "github.com/nu7hatch/gouuid"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	rand.Seed(time.Now().UnixNano())

	go gameStateProcessor()

	router = httprouter.New()

	uuid, _ := uuid.NewV4()

	os.Setenv("DICEKING_GENERATED_RESTART_KEY", uuid.String())

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
		return
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

	fmt.Printf("User Agent: %s\nRemote addr: %s\n", req.UserAgent(), req.RemoteAddr)

	fileBytes, err := ioutil.ReadFile("html/index.html")
	if err != nil {
		log.Print(err)
	}

	cookie := &http.Cookie{
		Name:   "RESTART_UUID",
		Value:  os.Getenv("DICEKING_GENERATED_RESTART_KEY"),
		Domain: req.URL.Hostname(),
	}
	http.SetCookie(w, cookie)

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
