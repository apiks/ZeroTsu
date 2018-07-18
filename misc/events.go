package misc

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
	"github.com/r-anime/ZeroTsu/config"
	"net/http"
	"strings"
	"time"
)

//Sets bot playing status and checks whether it's time to unban users
func StatusReady(s *discordgo.Session, e *discordgo.Ready) {
	s.UpdateStatus(0, "with her darling")

	//Checks whether it has to unban a user every 10 seconds
	for range time.NewTicker(10 * time.Second).C {

		//Saves current time
		t := time.Now()

		//Reads bannedUsers.json
		BannedUsersRead()

		//Goes through bannedUsers.json if it's not empty and unbans if needed
		if BannedUsersSlice != nil {
			for i := 0; i < len(BannedUsersSlice); i++ {
				difference := t.Sub(BannedUsersSlice[i].UnbanDate)
				if difference > 0 {

					MemberInfoRead()

					//Checks if user is in MemberInfo and assigns to user variable
					user, ok := MemberInfoMap[BannedUsersSlice[i].ID]
					if !ok {
						fmt.Print("User: " + BannedUsersSlice[i].User + " not found in memberInfo")
						return
					}

					UnbanUser(BannedUsersSlice[i].ID, s)

					//Sends a message to bot-log
					_, err := s.ChannelMessageSend(config.BotLogID, "User: "+user.Username+"#"+
						user.Discrim+" has been unbanned.")
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Removes the user ban from bannedUsers.json
					BannedUsersSlice = append(BannedUsersSlice[:i], BannedUsersSlice[i+1:]...)

					//Writes to bannedUsers.json
					BannedUsersWrite(BannedUsersSlice)
				}
			}
		}
	}
}

// Checks if it's time to send rss thread every 15 sec
func RssThreadReady(s *discordgo.Session, e *discordgo.Ready) {

	//Checks whether it has to post rss thread every 15 seconds
	for range time.NewTicker(15 * time.Second).C {

		RSSParser(*s)
	}

}

//Unbans a user via ID
func UnbanUser(id string, s *discordgo.Session) {

	//Reads memberInfo.json
	MemberInfoRead()

	//Unbans user
	err := s.GuildBanDelete(config.ServerID, id)
	if err != nil {

		fmt.Println("Error unbanning: ", err)
	}
}

// Pulls the rss thread and prints it
func RSSParser(s discordgo.Session) {

	// Pulls the feed from /r/anime and puts it in feed variable
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Transport: &UserAgentTransport{http.DefaultTransport}, Timeout: time.Second * 30}
	feed, err := fp.ParseURL("https://www.reddit.com/r/anime/new/.rss")
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Reads RssThreads files
	RssThreadsRead()
	RssThreadsCheckRead()

	//Saves current time
	t := time.Now()

	// Checks if the feed timeout to avoid error. If no nil then it continues down
	if feed != nil {

		// Removes a thread if more than 16 hours have passed
		for p := 0; p < len(ReadRssThreadsCheck); p++ {

			// Saves the date of removal in separate variable and then adds 10 hours to it
			tenHours := time.Hour * 10
			dateRemoval := ReadRssThreadsCheck[p].Date.Add(tenHours)

			// Calculates if it's time to remove
			difference := t.Sub(dateRemoval)
			if difference > 0 {

				// Removes the fact that the thread had been posted already
				RssThreadsCheckRemove(ReadRssThreadsCheck[p].Thread, ReadRssThreadsCheck[p].Date)
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
						}
					}

					if threadExists == false {

						// Writes to storage that the thread has been posted
						RssThreadsCheckWrite(ReadRssThreads[j].Thread, t)

						// Sends the thread to the channel
						_, err = s.ChannelMessageSend(ReadRssThreads[j].Channel, feed.Items[i].Link)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					}
				}
			}
		}
	}
}
