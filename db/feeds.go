package db

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"
)

// GetGuildFeeds returns the guild's feeds from in-memory
func GetGuildFeeds(guildID string) []entities.Feed {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetFeeds()
}

// SetGuildFeeds sets a target guild's feeds in-memory
func SetGuildFeeds(guildID string, feeds []entities.Feed) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].Lock()

	if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(feeds) >= 400 {
		entities.Guilds.Unlock()
		entities.Guilds.DB[guildID].Unlock()
		return fmt.Errorf("Error: You have reached the reddit feed autopost limit (400) for this server.")
	} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(feeds) >= 50 {
		entities.Guilds.Unlock()
		entities.Guilds.DB[guildID].Unlock()
		return fmt.Errorf("Error: You have reached the reddit feed autopost limit (50) for this server. Please remove some or increase them to 400 by upgrading to a premium server at <https://patreon.com/animeschedule>")
	}

	entities.Guilds.DB[guildID].SetFeeds(feeds)

	entities.Guilds.Unlock()
	entities.Guilds.DB[guildID].Unlock()

	entities.Guilds.DB[guildID].WriteData("rssThreads", entities.Guilds.DB[guildID].GetFeeds())
	return nil
}

// GetGuildFeed returns a guild's feed from in-memory
func GetGuildFeed(guildID, subreddit, channelID string) entities.Feed {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()
	entities.Guilds.DB[guildID].RLock()
	defer entities.Guilds.DB[guildID].RUnlock()

	for _, guildFeed := range entities.Guilds.DB[guildID].GetFeeds() {
		if guildFeed.GetSubreddit() == subreddit && guildFeed.GetChannelID() == channelID {
			return guildFeed
		}
	}

	return entities.Feed{}
}

// SetGuildFeed sets a guild's feed in-memory
func SetGuildFeed(guildID string, feed entities.Feed, delete ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	if len(delete) == 0 {
		if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetFeeds()) >= 400 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the reddit feed autopost limit (400) for this server.")
		} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetFeeds()) >= 50 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the reddit feed autopost limit (50) for this server. Please remove some or increase them to 400 by upgrading to a premium server at <https://patreon.com/animeschedule>")
		}
	}

	feed = feed.SetSubreddit(strings.ToLower(feed.GetSubreddit()))

	if len(delete) == 0 {
		var exists bool
		for _, guildFeed := range entities.Guilds.DB[guildID].GetFeeds() {
			if guildFeed.GetSubreddit() == feed.GetSubreddit() &&
				guildFeed.GetChannelID() == feed.GetChannelID() {
				exists = true
				break
			}
		}

		if !exists {
			entities.Guilds.DB[guildID].AppendToFeeds(feed)
		} else {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: That feed already exists.")
		}
	} else {
		err := deleteGuildFeed(guildID, feed)
		if err != nil {
			entities.Guilds.Unlock()
			return err
		}
	}
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("rssThreads", entities.Guilds.DB[guildID].GetFeeds())
	return nil
}

// deleteGuildFeed safely deletes a feed from the feeds slice
func deleteGuildFeed(guildID string, feed entities.Feed) error {
	var exists bool

	for i, guildFeed := range entities.Guilds.DB[guildID].GetFeeds() {
		if feed.GetSubreddit() == guildFeed.GetSubreddit() {
			if feed.GetChannelID() != "" && guildFeed.GetChannelID() != feed.GetChannelID() {
				continue
			}
			if feed.GetTitle() != "" && strings.ToLower(guildFeed.GetTitle()) != strings.ToLower(feed.GetTitle()) {
				continue
			}
			if feed.GetAuthor() != "" && strings.ToLower(guildFeed.GetAuthor()) != strings.ToLower(feed.GetAuthor()) {
				continue
			}
			if feed.GetPostType() != "" && strings.ToLower(guildFeed.GetPostType()) != strings.ToLower(feed.GetPostType()) {
				continue
			}

			entities.Guilds.DB[guildID].RemoveFromFeeds(i)
			exists = true
			break
		}
	}

	if !exists {
		return fmt.Errorf("Error: No such feed exists.")
	}

	return nil
}
