package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger/v2"
)

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

	if m.Author.ID == b.session.State.User.ID {
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

	log.Printf("Received message: guild <%s>, channel <%s>, user <%s>",
		guild.Name,
		channel.Name,
		m.Author.Username,
	)

	for _, u := range m.Mentions {
		if u.ID == b.session.State.User.ID {
			b.session.ChannelMessageSend(m.ChannelID, "Hi "+m.Author.Mention())
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
