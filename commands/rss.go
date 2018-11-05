package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Sets an RSS by author in the message channel
func setRssCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		author string
		thread string
	)

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "setrss OPTIONAL[/u/author] [thread name]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	if strings.Contains(commandStrings[1], "/u/") == true {
		author = commandStrings[1]
		thread = strings.Replace(messageLowercase, config.BotPrefix+"setrss "+commandStrings[1]+" ", "", 1)

	} else {
		// Removes the command from the string so we only have the set string which it'll check
		thread = strings.Replace(messageLowercase, config.BotPrefix+"setrss ", "", 1)
		author = "/u/AutoLovepon"
	}

	setRssThread(s, m, thread, author)
}

func setRssThread(s *discordgo.Session, m *discordgo.Message, thread string, author string) {

	threadExists, err := misc.RssThreadsWrite(thread, m.ChannelID, author)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	if threadExists == false {
		_, err := s.ChannelMessageSend(m.ChannelID, "`" + thread + "` has been added to the rss thread list.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "`" + thread + "` is already on the rss thread list.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	}
}

// Removes a previously set RSS
func removeRssCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		author string
		thread string
	)

	if len(misc.ReadRssThreads) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error. There are no set rss threads.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) == 1 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "removerss OPTIONAL[/u/author] [thread name]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	if strings.Contains(commandStrings[1], "/u/") == true {
		author = commandStrings[1]
		thread = strings.Replace(messageLowercase, config.BotPrefix+"removerss "+commandStrings[1]+" ", "", 1)
	} else {
		// Removes the command from the string so we only have the set string which it'll check
		thread = strings.Replace(messageLowercase, config.BotPrefix+"removerss ", "", 1)
		author = "/u/AutoLovepon"
	}

	// Calls the function to remove the threads from rssThreads.json
	threadExists, err := misc.RssThreadsRemove(thread, m.ChannelID, author)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	if threadExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "`" + thread + "` has been removed from the rss thread list.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Error: Thread does not exist in RSS list.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Prints all currently set RSS
func viewRssCommand(s *discordgo.Session, m *discordgo.Message) {

	var threads string

	if len(misc.ReadRssThreads) == 0 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set RSS threads.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Iterates through all the filters if they exist and adds them to the filters string and print them
	for i := 0; i < len(misc.ReadRssThreads); i++ {
		if len(threads) > 1850 {
			_, err := s.ChannelMessageSend(m.ChannelID, threads)
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					return
				}
			}
			threads = ""
		}

		if threads == "" {
			threads = "`" + misc.ReadRssThreads[i].Thread + " - " + misc.ReadRssThreads[i].Channel + " - " +
				misc.ReadRssThreads[i].Author + "`\n"
		} else {
			threads = threads + "\n `" + misc.ReadRssThreads[i].Thread + " - " + misc.ReadRssThreads[i].Channel + " - " +
				misc.ReadRssThreads[i].Author + "`\n"
		}
	}

	_, err := s.ChannelMessageSend(m.ChannelID, threads)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute:  setRssCommand,
		trigger:  "setrss",
		desc:     "Assigns an RSS to the channel.",
		elevated: true,
		category: "rss",
	})
	add(&command{
		execute:  removeRssCommand,
		trigger:  "removerss",
		desc:     "Removes a previously set RSS.",
		elevated: true,
		category: "rss",
	})
	add(&command{
		execute:  viewRssCommand,
		trigger:  "viewrss",
		aliases:  []string{"showrss", "rssview", "rssshow", "viewrs", "showrs"},
		desc:     "Prints all currently set RSS.",
		elevated: true,
		category: "rss",
	})
}