package entities

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

const DBPath = "database/guilds"

var (
	Guilds = NewGuildMap(make(map[string]*GuildInfo))

	guildFileNames = [...]string{"punishedUsers.json", "filters.json", "messReqs.json", "spoilerRoles.json", "rssThreads.json",
		"rssThreadCheck.json", "raffles.json", "waifus.json", "waifuTrades.json", "memberInfo.json", "emojiStats.json",
		"channelStats.json", "userChangeStats.json", "verifiedStats.json", "voteInfo.json", "tempCha.json",
		"reactJoin.json", "guildSettings.json", "autoposts.json"}
)

// GuildMap is a mutex-safe map of GuildInfo
type GuildMap struct {
	sync.RWMutex

	DB map[string]*GuildInfo
}

func NewGuildMap(DB map[string]*GuildInfo) *GuildMap {
	return &GuildMap{DB: DB}
}

// Init initializes a new guild with an empty GuildInfo Object
func (g *GuildMap) Init(guildID string) {
	path := fmt.Sprintf("%s/%s", DBPath, guildID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0777)
		if err != nil {
			log.Println(err)
			return
		}
	}

	for _, name := range guildFileNames {
		file, err := os.OpenFile(fmt.Sprintf("%s/%s/%s", DBPath, guildID, name), os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err)
			continue
		}
		err = file.Close()
		if err != nil {
			log.Println(err)
			continue
		}
	}

	g.Lock()
	defer g.Unlock()
	g.DB[guildID] = &GuildInfo{
		ID: guildID,
		GuildSettings: &GuildSettings{
			Prefix:              ".",
			ReactsModule:        true,
			PingMessage:         "Hmmm~ So this is what you do all day long?",
			BotLog:              &Cha{},
			VoteChannelCategory: &Cha{},
			OptInUnder:          &Role{},
			OptInAbove:          &Role{},
			MutedRole:           &Role{},
		},
		MemberInfoMap:   make(map[string]*UserInfo),
		SpoilerMap:      make(map[string]*discordgo.Role),
		EmojiStats:      make(map[string]*Emoji),
		ChannelStats:    make(map[string]*Channel),
		UserChangeStats: make(map[string]int),
		VerifiedStats:   make(map[string]int),
		VoteInfoMap:     make(map[string]*VoteInfo),
		TempChaMap:      make(map[string]*TempChaInfo),
		ReactJoinMap:    make(map[string]*ReactJoin),
		ExtensionList:   make(map[string]string),
		Autoposts:       make(map[string]*Cha),
	}
}

// Load loads a preexisting guild
func (g *GuildMap) Load(guildID string) error {
	g.Init(guildID)

	g.RLock()
	guild := g.DB[guildID]
	g.RUnlock()

	files, err := IOReadDir(fmt.Sprintf("%s/%s/", DBPath, guildID))
	if err != nil {
		return err
	}

	// Load guild settings first because some files check against bools in the settings
	err = guild.Load("guildSettings.json", guildID)
	if err != nil {
		return err
	}

	// Load each of the guild files
	for _, file := range files {
		if file == "guildSettings.json" {
			continue
		}
		err = guild.Load(file, guildID)
		if err != nil {
			return err
		}
	}

	// Init default settings
	if _, ok := g.DB[guildID].Autoposts["newepisodes"]; ok {
		Mutex.Lock()
		SetupGuildSub(guildID)
		Mutex.Unlock()
	}

	g.Lock()
	g.DB[guildID] = guild
	g.Unlock()

	return nil
}

// LoadAll loads all guilds from storage
func (g *GuildMap) LoadAll() {
	if _, err := os.Stat("database"); os.IsNotExist(err) {
		err := os.Mkdir("database", 0777)
		if err != nil {
			log.Println(err)
			return
		}
	}
	if _, err := os.Stat(DBPath); os.IsNotExist(err) {
		err := os.Mkdir(DBPath, 0777)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	folders, err := ioutil.ReadDir(DBPath)
	if err != nil {
		log.Panicln(err)
	}

	for _, f := range folders {
		if !f.IsDir() {
			continue
		}
		err = g.Load(f.Name())
		if err != nil {
			log.Panicln(err)
		}
	}
}

// HandleNewGuild initializes a guild if it's not in memory
func HandleNewGuild(guildID string) {
	Guilds.RLock()
	if _, ok := Guilds.DB[guildID]; !ok {
		Guilds.RUnlock()
		Guilds.Init(guildID)
		return
	}
	Guilds.RUnlock()
}
