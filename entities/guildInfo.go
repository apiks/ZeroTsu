package entities

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/sasha-s/go-deadlock"
)

// GuildInfo contains all the data a guild can contain
type GuildInfo struct {
	deadlock.RWMutex

	ID            string
	GuildSettings GuildSettings
	Feeds         []Feed
	FeedChecks    []FeedCheck
	Raffles       []*Raffle
	ReactJoinMap  map[string]*ReactJoin
	Autoposts     map[string]Cha
}

func (g *GuildInfo) SetID(id string) {
	g.Lock()
	g.ID = id
	g.Unlock()
}

func (g *GuildInfo) GetID() string {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return ""
	}
	return g.ID
}

func (g *GuildInfo) SetGuildSettings(guildSettings GuildSettings) {
	g.Lock()
	g.GuildSettings = guildSettings
	g.Unlock()
}

func (g *GuildInfo) GetGuildSettings() GuildSettings {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return GuildSettings{}
	}
	return g.GuildSettings
}

func (g *GuildInfo) AssignToFeeds(index int, feed Feed) {
	g.Lock()
	g.Feeds[index] = feed
	g.Unlock()
}

func (g *GuildInfo) AppendToFeeds(feed Feed) {
	g.Lock()
	g.Feeds = append(g.Feeds, feed)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromFeeds(index int) {
	g.Lock()
	g.Feeds = append(g.Feeds[:index], g.Feeds[index+1:]...)
	g.Unlock()
}

func (g *GuildInfo) SetFeeds(feeds []Feed) {
	g.Lock()
	g.Feeds = feeds
	g.Unlock()
}

func (g *GuildInfo) GetFeeds() []Feed {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.Feeds
}

func (g *GuildInfo) AssignToFeedChecks(index int, feedCheck FeedCheck) {
	g.Lock()
	g.FeedChecks[index] = feedCheck
	g.Unlock()
}

func (g *GuildInfo) AppendToFeedChecks(feedCheck FeedCheck) {
	g.Lock()
	g.FeedChecks = append(g.FeedChecks, feedCheck)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromFeedChecks(index int) {
	g.Lock()
	g.FeedChecks = append(g.FeedChecks[:index], g.FeedChecks[index+1:]...)
	g.Unlock()
}

func (g *GuildInfo) SetFeedChecks(feedChecks []FeedCheck) {
	g.Lock()
	g.FeedChecks = feedChecks
	g.Unlock()
}

func (g *GuildInfo) GetFeedChecks() []FeedCheck {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.FeedChecks
}

func (g *GuildInfo) AppendToRaffles(raffle *Raffle) {
	g.Lock()
	g.Raffles = append(g.Raffles, raffle)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromRaffles(index int) {
	g.Lock()
	if index < len(g.Raffles)-1 {
		copy(g.Raffles[index:], g.Raffles[index+1:])
	}
	g.Raffles[len(g.Raffles)-1] = nil
	g.Raffles = g.Raffles[:len(g.Raffles)-1]
	g.Unlock()
}

func (g *GuildInfo) SetRaffles(raffles []*Raffle) {
	g.Lock()
	g.Raffles = raffles
	g.Unlock()
}

func (g *GuildInfo) GetRaffles() []*Raffle {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.Raffles
}

func (g *GuildInfo) AssignToReactJoinMap(key string, reactJoin *ReactJoin) {
	g.Lock()
	g.ReactJoinMap[key] = reactJoin
	g.Unlock()
}

func (g *GuildInfo) RemoveFromReactJoinMap(key string) {
	g.Lock()
	delete(g.ReactJoinMap, key)
	g.Unlock()
}

func (g *GuildInfo) SetReactJoinMap(reactJoinMap map[string]*ReactJoin) {
	g.Lock()
	g.ReactJoinMap = reactJoinMap
	g.Unlock()
}

func (g *GuildInfo) GetReactJoinMap() map[string]*ReactJoin {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.ReactJoinMap
}

func (g *GuildInfo) AssignToAutoposts(key string, autopost Cha) {
	g.Lock()
	g.Autoposts[key] = autopost
	g.Unlock()
}

func (g *GuildInfo) RemoveFromAutoposts(key string) {
	g.Lock()
	delete(g.Autoposts, key)
	g.Unlock()
}

func (g *GuildInfo) SetAutoposts(autoposts map[string]Cha) {
	g.Lock()
	g.Autoposts = autoposts
	g.Unlock()
}

func (g *GuildInfo) GetAutoposts() map[string]Cha {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.Autoposts
}

// Load loads a guild file into the guild memory
func (g *GuildInfo) Load(file, guildID string) error {
	fileData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%s", DBPath, guildID, file))
	if err != nil {
		return err
	}
	if len(fileData) == 0 {
		return nil
	}

	g.Lock()
	defer g.Unlock()

	switch file {
	case "guildSettings.json":
		return json.Unmarshal(fileData, &g.GuildSettings)
	case "rssThreads.json":
		return json.Unmarshal(fileData, &g.Feeds)
	case "rssThreadCheck.json":
		return json.Unmarshal(fileData, &g.FeedChecks)
	case "raffles.json":
		return json.Unmarshal(fileData, &g.Raffles)
	case "reactJoin.json":
		return json.Unmarshal(fileData, &g.ReactJoinMap)
	case "autoposts.json":
		return json.Unmarshal(fileData, &g.Autoposts)
	}

	return nil
}

// WriteData writes some kind of guild data to the target guild file
func (g *GuildInfo) WriteData(fileName string, data interface{}) {
	marshaledData, err := json.MarshalIndent(&data, "", "    ")
	if err != nil {
		log.Println(err)
		return
	}
	if len(marshaledData) == 0 {
		return
	}

	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%s/%s.json", g.GetID(), fileName), marshaledData, 0644)
	if err != nil {
		log.Println("WriteData error:", err)
		return
	}
}
