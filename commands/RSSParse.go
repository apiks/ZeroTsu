package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Sets an RSS for a channel that it'll check and post in that channel
func RSSHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Checks if it's within the /r/anime server
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, err = s.Channel(m.ChannelID)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
	if ch.GuildID == config.ServerID {

		//Puts the command to lowercase
		messageLowercase := strings.ToLower(m.Content)

		//Checks if BotPrefix + setrss was used or BotPrefix + viewrss or BotPrefix + deleterss
		if strings.HasPrefix(messageLowercase, config.BotPrefix+"setrss ") && (messageLowercase != (config.BotPrefix + "setrss")) {

			if m.Author.ID == config.BotID {
				return
			}

			var (
				author string
				thread string
			)

			messageSplit := strings.Split(messageLowercase, " ")

			if strings.Contains(messageSplit[1], "/u/") == true {

				author = messageSplit[1]
				thread = strings.Replace(messageLowercase, config.BotPrefix+"setrss "+messageSplit[1]+" ", "", 1)

			} else {

				// Removes the command from the string so we only have the set string which it'll check
				thread = strings.Replace(messageLowercase, config.BotPrefix+"setrss ", "", 1)
				author = "/u/AutoLovepon"
			}

			setRssThread(*s, *m, thread, author)

		} else if strings.HasPrefix(messageLowercase, config.BotPrefix+"viewrss") {

			if m.Author.ID == config.BotID {
				return
			}

			//Reads all the rss threads from rssThreads.json
			misc.RssThreadsRead()

			//Creates a string variable to store the threads in for showing later
			var threads string

			//Iterates through all the filters if they exist and adds them to the filters string
			if len(misc.ReadRssThreads) != 0 {

				for i := 0; i < len(misc.ReadRssThreads); i++ {

					if len(threads) > 1850 {

						//If there are no rss threads give error, else print the threads.
						if len(misc.ReadRssThreads) == 0 {

							// Sets and prints error message
							failure := "Error. There are no set rss threads."
							_, err := s.ChannelMessageSend(m.ChannelID, failure)
							if err != nil {

								fmt.Println("Error: ", err)
							}
						} else {

							_, err := s.ChannelMessageSend(m.ChannelID, threads)
							if err != nil {

								fmt.Println("Error: ", err)
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
			}

			//If there are no rss threads give error, else print the threads.
			if len(misc.ReadRssThreads) == 0 {

				// Sets and prints error message
				failure := "Error. There are no set rss threads."
				_, err := s.ChannelMessageSend(m.ChannelID, failure)
				if err != nil {

					fmt.Println("Error: ", err)
				}
			} else {

				_, err := s.ChannelMessageSend(m.ChannelID, threads)
				if err != nil {

					fmt.Println("Error: ", err)
				}
			}

		} else if strings.HasPrefix(messageLowercase, config.BotPrefix+"removerss ") && (messageLowercase != (config.BotPrefix + "removerss")) {

			if m.Author.ID == config.BotID {
				return
			}

			var (
				author string
				thread string
			)

			//Reads all the rss threads from rssThreads.json
			misc.RssThreadsRead()

			messageSplit := strings.Split(messageLowercase, " ")

			if strings.Contains(messageSplit[1], "/u/") == true {

				author = messageSplit[1]
				thread = strings.Replace(messageLowercase, config.BotPrefix+"setrss "+messageSplit[1]+" ", "", 1)

			} else {

				// Removes the command from the string so we only have the set string which it'll check
				thread = strings.Replace(messageLowercase, config.BotPrefix+"removerss ", "", 1)
				author = "/u/AutoLovepon"
			}

			//Checks if there's any rss threads, else prints success.
			if len(misc.ReadRssThreads) == 0 {

				// Sets and prints error message
				failure := "Error. There are no set rss threads."
				_, err := s.ChannelMessageSend(m.ChannelID, failure)
				if err != nil {

					fmt.Println("Error: ", err)
				}
			} else {

				//Calls the function to remove the threads from rssThreads.json
				misc.RssThreadsRemove(thread, m.ChannelID, author)

				//Prints success
				success := "`" + thread + "` has been removed from the rss thread list."
				_, err := s.ChannelMessageSend(m.ChannelID, success)
				if err != nil {

					fmt.Println("Error: ", err)
				}
			}
		}
	}
}

func setRssThread(s discordgo.Session, m discordgo.MessageCreate, thread string, author string) {

	misc.RssThreadsWrite(thread, m.ChannelID, author)

	if misc.ThreadExists == false {

		//Prints success
		success := "`" + thread + "` has been added to the rss thread list."
		_, err := s.ChannelMessageSend(m.ChannelID, success)
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {

		//Prints failure
		failure := "`" + thread + "` is already on the rss thread list."
		_, err := s.ChannelMessageSend(m.ChannelID, failure)
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}
