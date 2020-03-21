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
	dbPath := flag.String("db", "", "db")
	flag.Parse()

	if *token == "" {
		log.Fatal("token is empty")
	}
	if *dbPath == "" {
		log.Fatal("db is empty")
	}

	log.Print("Starting")
	b, err := bot.New(*token, *dbPath)
	if err != nil {
		log.Fatal(err)
	}

	err = b.Open()
	if err != nil {
		log.Fatal(err)
	}
	b.Listen()
	log.Print("Connected")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Print("Stopping")
	err = b.Close()
	if err != nil {
		log.Fatal(err)
	}
}
