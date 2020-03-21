package bot

import (
	"log"
	"math"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger/v2"
)

// 1 in 1000
const respProbThreshold = math.MaxUint64 / 1000

// Bot represents a Markov bot with a Discord session and a database.
type Bot struct {
	session *discordgo.Session
	dbOpts  badger.Options
	db      *badger.DB
}

// New creates and configures a Bot.
func New(token string, dbPath string) (*Bot, error) {
	dbOpts := badger.DefaultOptions(dbPath)

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	b := &Bot{s, dbOpts, nil}

	s.AddHandler(func(s2 *discordgo.Session, m *discordgo.MessageCreate) {
		if s2 != b.session {
			log.Fatal("Session seems to have changed")
		}

		b.handleMessageCreate(m)
	})

	return b, nil
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

// Open opens the database and creates a websocket connection to Discord.
func (b *Bot) Open() error {
	db, err := badger.Open(b.dbOpts)
	if err != nil {
		b.session.Close()
		return err
	}
	b.db = db

	err = b.session.Open()
	if err != nil {
		return err
	}

	return nil
}

// Close closes the database and the websocket connection to Discord.
func (b *Bot) Close() error {
	sErr := b.session.Close()
	dbErr := b.db.Close()
	// The dbErr is more important so return that one if it exists.
	if dbErr != nil {
		return dbErr
	}
	return sErr
}

func (b *Bot) registerMessage(guildID, msg string) error {
	return nil
}

func (b *Bot) generateMessage(guildID string) string {
	return "Beep boop"
}
