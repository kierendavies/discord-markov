package bot

import (
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger/v2"
)

// 1 in 1000
const respProbThreshold uint64 = math.MaxUint64 / 1000

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

// Listen starts the bot processing and responding to new messages. The return
// value is a function which, when called, stops the listening.
func (b *Bot) Listen() func() {
	return b.session.AddHandler(func(s2 *discordgo.Session, m *discordgo.MessageCreate) {
		if s2 != b.session {
			log.Fatal("Session seems to have changed")
		}

		b.handleMessageCreate(m)
	})
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
			log.Printf("Error: %+v", err)
		}

		return
	}

	// Ignore messages with no text content, e.g. just images.
	if m.Content == "" {
		return
	}

	guild, err := b.session.Guild(m.GuildID)
	if err != nil {
		log.Printf("Error: %+v", err)
		return
	}

	channel, err := b.session.Channel(m.ChannelID)
	if err != nil {
		log.Printf("Error: %+v", err)
		return
	}

	log.Printf("Received message from %s/#%s/%s#%s",
		guild.Name,
		channel.Name,
		m.Author.Username,
		m.Author.Discriminator,
	)

	content, err := m.ContentWithMoreMentionsReplaced(b.session)
	if err != nil {
		log.Printf("Error: %+v", err)
		return
	}

	err = b.registerMessage(m.GuildID, content)
	if err != nil {
		log.Printf("Error: %+v", err)
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
		log.Printf("Sending message to %s/#%s",
			guild.Name,
			channel.Name,
		)

		resp, err := b.generateMessage(m.GuildID)
		if err != nil {
			log.Printf("Error: %+v", err)
			return
		}

		_, err = b.session.ChannelMessageSend(m.ChannelID, resp)
		if err != nil {
			log.Printf("Error: %+v", err)
			return
		}
	}
}

func (b *Bot) registerMessage(guildID, msg string) error {
	chains := tokenChains(msg)
	keys := make([]string, 0, len(chains))
	for _, ts := range chains {
		keys = append(keys, guildID+":"+ts)
	}
	err := b.incrementCounts(keys)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) generateMessage(guildID string) (string, error) {
	tokens := make([]string, 0)
	tokens = append(tokens, stx)

	for tokens[len(tokens)-1] != etx {
		chainLen := chainLenGen
		if chainLen > len(tokens) {
			chainLen = len(tokens)
		}
		chain := strings.Join(tokens[len(tokens)-chainLen:len(tokens)], tokenSeparator)

		counts, err := b.getCounts(guildID + ":" + chain)
		if err != nil {
			return "", err
		}
		nextToken := weightedChoice(counts)
		tokens = append(tokens, nextToken)
	}

	return strings.Join(tokens[1:len(tokens)-1], tokenSeparator), nil
}
