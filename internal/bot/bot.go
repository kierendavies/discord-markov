package bot

import (
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger/v2"
)

// 1 in 1000
const respProbThreshold = math.MaxUint64 / 1000

const dbGCInterval = 5 * time.Minute
const dbGCDiscardRatio = 0.5

// Bot represents a Markov bot with a Discord session and a database.
type Bot struct {
	session    *discordgo.Session
	dbOpts     badger.Options
	db         *badger.DB
	dbGCTicker *time.Ticker
	stopCh     chan struct{}
	waitGroup  *sync.WaitGroup
}

// New creates and configures a Bot.
func New(token string, dbPath string) (*Bot, error) {
	dbOpts := badger.DefaultOptions(dbPath)

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	b := &Bot{
		session:   s,
		dbOpts:    dbOpts,
		waitGroup: &sync.WaitGroup{},
	}

	s.AddHandler(func(s2 *discordgo.Session, m *discordgo.MessageCreate) {
		if s2 != b.session {
			log.Fatal("Session seems to have changed")
		}

		b.handleMessageCreate(m)
	})

	return b, nil
}

// Open opens the database and creates a websocket connection to Discord.
func (b *Bot) Open() error {
	b.stopCh = make(chan struct{})

	db, err := badger.Open(b.dbOpts)
	if err != nil {
		b.session.Close()
		return err
	}
	b.db = db

	b.dbGCTicker = time.NewTicker(dbGCInterval)
	b.waitGroup.Add(1)
	go func() {
		defer b.waitGroup.Done()
		for {
			select {
			case <-b.dbGCTicker.C:
				log.Print("Running database garbage collection")
				b.db.RunValueLogGC(dbGCDiscardRatio)
			case <-b.stopCh:
				return
			}
		}
	}()

	err = b.session.Open()
	if err != nil {
		return err
	}

	return nil
}

// Close closes the database and the websocket connection to Discord.
func (b *Bot) Close() error {
	b.dbGCTicker.Stop()
	close(b.stopCh)
	b.waitGroup.Wait()

	sErr := b.session.Close()
	dbErr := b.db.Close()
	// The dbErr is more important so return that one if it exists.
	if dbErr != nil {
		return dbErr
	}
	return sErr
}

func (b *Bot) handleMessageCreate(m *discordgo.MessageCreate) {
	var err error

	// Ignore my own messages.
	if m.Author.ID == b.session.State.User.ID {
		return
	}

	// Politely reject direct messages.
	if m.GuildID == "" {
		log.Printf("Ignoring DM from %s#%s", m.Author.Username, m.Author.Discriminator)

		_, err = b.session.ChannelMessageSend(m.ChannelID, "Sorry, I can't process direct messages")
		if err != nil {
			log.Print(err)
		}

		return
	}

	// Ignore messages with no text content, e.g. just images.
	if m.Content == "" {
		return
	}

	guild, err := b.session.Guild(m.GuildID)
	if err != nil {
		log.Print(err)
		return
	}

	channel, err := b.session.Channel(m.ChannelID)
	if err != nil {
		log.Print(err)
		return
	}

	log.Printf("Received message from %s/%s/%s#%s",
		guild.Name,
		channel.Name,
		m.Author.Username,
		m.Author.Discriminator,
	)

	err = b.registerMessage(m.GuildID, m.Content)
	if err != nil {
		log.Print(err)
		return
	}

	shouldRespond := false

	// Respond to mentions.
	for _, u := range m.Mentions {
		if u.ID == b.session.State.User.ID {
			shouldRespond = true
			break
		}
	}

	// Randomly respond sometimes.
	if !shouldRespond {
		if rand.Uint64() <= respProbThreshold {
			shouldRespond = true
		}
	}

	if shouldRespond {
		log.Printf("Sending message to %s/%s",
			guild.Name,
			channel.Name,
		)

		resp := b.generateMessage(m.GuildID)

		_, err = b.session.ChannelMessageSend(m.ChannelID, resp)
		if err != nil {
			log.Print(err)
			return
		}
	}
}

func (b *Bot) registerMessage(guildID, msg string) error {
	return nil
}

func (b *Bot) generateMessage(guildID string) string {
	return "Beep boop"
}
