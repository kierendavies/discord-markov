package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
)

const chainLen = 3
const (
	stx = '\x02'
	etx = '\x03'
)

func main() {
	token := flag.String("token", "", "token")
	guildID := flag.String("guild", "", "guild")
	flag.Parse()
	_ = token
	_ = guildID

	// fetch(*token, *guildID)
	process()
}

func fetch(token, guildID string) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		panic(err)
	}
	s.State.GuildAdd(guild)

	channels, err := s.GuildChannels(guildID)
	if err != nil {
		panic(err)
	}
	for _, channel := range channels {
		err = s.State.ChannelAdd(channel)
		if err != nil {
			panic(err)
		}
	}

	messages := make([]string, 0)

	for _, channel := range channels {
		if channel.Type != discordgo.ChannelTypeGuildText {
			continue
		}

		fmt.Println(channel.Name)

		beforeID := ""
		for {
			chMsgs, err := s.ChannelMessages(channel.ID, 100, beforeID, "", "")
			if err != nil {
				if restErr, ok := err.(*discordgo.RESTError); ok {
					if restErr.Response.StatusCode == http.StatusForbidden {
						break
					}
				}
				panic(err)
			}

			if len(chMsgs) == 0 {
				break
			}

			for _, m := range chMsgs {
				content, err := m.ContentWithMoreMentionsReplaced(s)
				if err != nil {
					panic(err)
				}
				messages = append(messages, content)
			}

			beforeID = chMsgs[len(chMsgs)-1].ID

			fmt.Println(chMsgs[0].Timestamp)
		}
	}

	f, err := os.Create("messages.json")
	defer f.Close()
	if err != nil {
		panic(err)
	}

	err = json.NewEncoder(f).Encode(messages)
	if err != nil {
		panic(err)
	}
}

func process() {
	f, err := os.Open("messages.json")
	defer f.Close()
	if err != nil {
		panic(err)
	}

	var messages []string
	err = json.NewDecoder(f).Decode(&messages)
	if err != nil {
		panic(err)
	}

	tr := make(map[string]map[rune]uint64)

	for _, m := range messages {
		chain := make([]rune, 0)
		for i := 0; i < chainLen; i++ {
			chain = append(chain, stx)
		}

		for _, c := range m {
			if _, ok := tr[string(chain)]; !ok {
				tr[string(chain)] = make(map[rune]uint64)
			}
			tr[string(chain)][c]++
			chain = append(chain[1:], c)
		}

		if _, ok := tr[string(chain)]; !ok {
			tr[string(chain)] = make(map[rune]uint64)
		}
		tr[string(chain)][etx]++
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		m := make([]rune, 0)
		for i := 0; i < chainLen; i++ {
			m = append(m, stx)
		}

		for m[len(m)-1] != etx {
			m = append(m, weightedChoice(tr[string(m[len(m)-chainLen:])]))
		}

		fmt.Println(string(m[chainLen : len(m)-1]))
		reader.ReadLine()
	}
}

func pp(x interface{}) {
	b, err := json.MarshalIndent(x, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func weightedChoice(choices map[rune]uint64) rune {
	if len(choices) == 0 {
		panic("choices is empty")
	}

	keys := make([]rune, 0, len(choices))
	cumsums := make([]uint64, 0, len(choices))
	var sum uint64 = 0

	for k, v := range choices {
		keys = append(keys, k)
		sum += v
		cumsums = append(cumsums, sum)
	}

	r := uint64(rand.Int63n(int64(sum)))
	for i, cs := range cumsums {
		if r < cs {
			return keys[i]
		}
	}

	panic("something went wrong with weightedChoice")
}
