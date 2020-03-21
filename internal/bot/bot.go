package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func New(token string) (*discordgo.Session, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	s.AddHandler(handleMessageCreate)

	return s, nil
}

func handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println(m)

	if m.Author.ID == s.State.User.ID {
		return
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Print(err)
		return
	}

	channel, err := s.Channel(m.ChannelID)
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
		if u.ID == s.State.User.ID {
			s.ChannelMessageSend(m.ChannelID, "Hi "+m.Author.Mention())
			return
		}
	}
}
