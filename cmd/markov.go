package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kierendavies/discord-markov/internal/bot"
)

func main() {
	token := flag.String("token", "", "token")
	flag.Parse()

	if *token == "" {
		log.Fatal("Token is empty")
	}

	log.Print("Starting")
	b, err := bot.New(*token)
	if err != nil {
		log.Fatal(err)
	}

	err = b.Open()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Connected")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Print("Stopping")
	b.Close()
}
