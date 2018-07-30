package misc

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"

	"github.com/r-anime/ZeroTsu/config"
)

// Periodic events such as Unbanning and RSS timer every 15 seconds
func StatusReady(s *discordgo.Session, e *discordgo.Ready) {

	// Saves program from panic and continues running normally without executing the command if it happens
	//defer func() {
	//	if rec := recover(); rec != nil {
	//		_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
	//		if err != nil {
	//
	//			fmt.Println(err.Error())
	//			fmt.Println(rec)
	//		}
	//	}
	//}()

	err := s.UpdateStatus(0, "with her darling")
	if err != nil {

		_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {

			fmt.Println(err.Error())
		}
	}

	for range time.NewTicker(30 * time.Second).C {

		// Checks whether it has to post rss thread every 15 seconds
		RSSParser(s)

		// Goes through bannedUsers.json if it's not empty and unbans if needed
		//if BannedUsersSlice != nil {
		//	if len(BannedUsersSlice) != 0 {
		//
		//		t := time.Now()
		//
		//		for i := 0; i < len(BannedUsersSlice); i++ {
		//			difference := t.Sub(BannedUsersSlice[i].UnbanDate)
		//			if difference > 0 {
		//
		//				// Checks if user is in MemberInfo and assigns to user variable if true
		//				user, ok := MemberInfoMap[BannedUsersSlice[i].ID]
		//				if !ok {
		//					continue
		//				}
		//
		//				// Unbans user
		//				err := s.GuildBanDelete(config.ServerID, BannedUsersSlice[i].ID)
		//				if err != nil {
		//
		//					_, _ = s.ChannelMessageSend(config.BotLogID, err.Error())
		//				}
		//
		//				// Sends a message to bot-log
		//				_, _ = s.ChannelMessageSend(config.BotLogID, "User: " + user.Username + "#"+
		//					user.Discrim+ " has been unbanned.")
		//
		//				// Removes the user ban from bannedUsers.json
		//				BannedUsersSlice = append(BannedUsersSlice[:i], BannedUsersSlice[i+1:]...)
		//
		//				// Writes to bannedUsers.json
		//				BannedUsersWrite(BannedUsersSlice)
		//			}
		//		}
		//	}
		//}
	}
}

// Pulls the rss thread and prints it
func RSSParser(s *discordgo.Session) {

	// Pulls the feed from /r/anime and puts it in feed variable
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Transport: &UserAgentTransport{http.DefaultTransport}, Timeout: time.Minute * 3}
	feed, err := fp.ParseURL("http://www.reddit.com/r/anime/new/.rss")
	if err != nil {

		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + ". Feed var error.")
		if err != nil {

			return
		}
		return
	}

	t := time.Now()

	// Removes a thread if more than 16 hours have passed
	for p := 0; p < len(ReadRssThreadsCheck); p++ {

		// Saves the date of removal in separate variable and then adds 10 hours to it
		tenHours := time.Hour * 10
		dateRemoval := ReadRssThreadsCheck[p].Date.Add(tenHours)

		// Calculates if it's time to remove
		difference := t.Sub(dateRemoval)
		if difference > 0 {

			// Removes the fact that the thread had been posted already
			RssThreadsTimerRemove(ReadRssThreadsCheck[p].Thread, ReadRssThreadsCheck[p].Date)
		}
	}

	// Iterates through each feed item to see if it finds something from storage
	for i := 0; i < len(feed.Items); i++ {
		for j := 0; j < len(ReadRssThreads); j++ {

			itemTitleLowercase := strings.ToLower(feed.Items[i].Title)
			itemAuthorLowercase := strings.ToLower(feed.Items[i].Author.Name)
			storageAuthorLowercase := strings.ToLower(ReadRssThreads[j].Author)

			if strings.Contains(itemTitleLowercase, ReadRssThreads[j].Thread) &&
				strings.Contains(itemAuthorLowercase, storageAuthorLowercase) {

				threadExists := false

				for k := 0; k < len(ReadRssThreadsCheck); k++ {
					if ReadRssThreadsCheck[k].Thread == ReadRssThreads[j].Thread {

						threadExists = true
						break
					}
				}

				if threadExists != false {

					return
				}

				// Writes to storage that the thread has been posted
				RssThreadsTimerWrite(ReadRssThreads[j].Thread, t)

				_, err = s.ChannelMessageSend(ReadRssThreads[j].Channel, feed.Items[i].Link)
				if err != nil {

					_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
					if err != nil {

						return
					}
				}
			}
		}
	}
}