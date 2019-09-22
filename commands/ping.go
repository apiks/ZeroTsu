package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"time"

	"github.com/r-anime/ZeroTsu/misc"
)

// Prints a message to see if the BOT is alive
func pingCommand(s *discordgo.Session, m *discordgo.Message) {
	misc.MapMutex.Lock()
	err := pingEmbed(s, m)
	if err != nil {
		if m.GuildID != "" {
			guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
			misc.CommandErrorHandler(s, m, err, guildBotLog)
		}
	}
	misc.MapMutex.Unlock()
}

func pingEmbed(s *discordgo.Session, m *discordgo.Message) error {
	embed := &discordgo.MessageEmbed{
		Title:       ":ping_pong:",
		Description: fmt.Sprintf("\n%v", misc.GuildMap[m.GuildID].GuildConfig.PingMessage),
		Color:       16758465,
		Thumbnail: &discordgo.MessageEmbedThumbnail {
			URL:s.State.User.AvatarURL("256"),
		},
	}

	// Parses and edits message with how long it took to send message
	now := time.Now()
	embedMsg, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		return err
	}
	delay, err := embedMsg.Timestamp.Parse()
	if err != nil {
		log.Println(err)
		return nil
	}
	difference := delay.Sub(now).Truncate(time.Millisecond).String()
	embed = &discordgo.MessageEmbed{
		Title:       fmt.Sprintf(":ping_pong: %v", difference),
		Description: fmt.Sprintf("%v", misc.GuildMap[m.GuildID].GuildConfig.PingMessage),
		Color:       16758465,
		Thumbnail: &discordgo.MessageEmbedThumbnail {
			URL:s.State.User.AvatarURL("256"),
		},
	}
	_, err = s.ChannelMessageEditEmbed(embedMsg.ChannelID, embedMsg.ID, embed)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	add(&command{
		execute:  pingCommand,
		trigger:  "ping",
		aliases:  []string{"pingme"},
		desc:     "See if I respond and how fast",
		elevated: true,
		category: "misc",
	})
}
