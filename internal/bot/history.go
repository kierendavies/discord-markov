package bot

import "log"

// ProcessHistory fetches and processes the message history from a given channel.
func (b *Bot) ProcessHistory(channelID string) error {
	channel, err := b.session.Channel(channelID)
	if err != nil {
		return err
	}

	oldestID := ""
	msgCount := 0
	for {
		msgs, err := b.session.ChannelMessages(channelID, 100, oldestID, "", "")
		if err != nil {
			return err
		}
		if len(msgs) == 0 {
			break
		}
		for _, msg := range msgs {
			// Ignore my own messages.
			if m.Author.ID == b.session.State.User.ID {
				continue
			}

			content, err := msg.ContentWithMoreMentionsReplaced(b.session)
			if err != nil {
				return err
			}
			b.registerMessage(channel.GuildID, content)
		}

		msgCount += len(msgs)
		log.Printf("Processed %d messages back to %s", msgCount, msgs[len(msgs)-1].Timestamp)

		oldestID = msgs[len(msgs)-1].ID
	}

	return nil
}
