package main

import (
	"flag"
	"log"

	"github.com/kierendavies/discord-markov/internal/bot"
)

func main() {
	token := flag.String("token", "", "token")
	dbPath := flag.String("db", "", "db")
	channelID := flag.String("channelID", "", "channelID")
	flag.Parse()

	if *token == "" {
		log.Fatal("token is empty")
	}
	if *dbPath == "" {
		log.Fatal("db is empty")
	}
	if *channelID == "" {
		log.Fatal("channelID is empty")
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
	defer b.Close()
	log.Print("Connected")

	err = b.ProcessHistory(*channelID)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Stopping")
	err = b.Close()
	if err != nil {
		log.Fatal(err)
	}
}
