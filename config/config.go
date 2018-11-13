package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// File for Bot, server, channel and role info

var (
	Token        			string
	BotPrefix    			string
	BotID        			string
	ServerID     			string
	BotLogID     			string
	CommandRoles 			[]string
	OptInUnder   			string
	OptInAbove   			string
	VoiceChaID   			string
	Website					string
	ModCategoryID			string
	VoteChannelCategoryID 	string

	RedditAppName			string
	RedditAppSecret			string
	DiscordAppSecret		string

	config 					*configStruct
	configsecrets 			*configSecrets
)

type configStruct struct {
	Token        			string   		`json:"-"`
	BotPrefix    			string   		`json:"BotPrefix"`
	BotID        			string   		`json:"BotID"`
	ServerID     			string   		`json:"ServerID"`
	BotLogID     			string   		`json:"BotLogID"`
	CommandRoles 			[]string 		`json:"CommandRoles"`
	OptInUnder   			string   		`json:"OptInUnder"`
	OptInAbove   			string   		`json:"OptInAbove"`
	VoiceChaID   			string   		`json:"VoiceChaID"`
	Website		 			string	 		`json:"Website"`
	ModCategoryID			string	 		`json:"ModCategoryID"`
	VoteChannelCategoryID 	string 	`json:"VoteChannelCategoryID"`
}

type configSecrets struct {
	RedditAppName			string	`json:"RedditName"`
	RedditAppSecret			string	`json:"RedditSecret"`
	DiscordAppSecret		string	`json:"DiscordSecret"`
}

// Loads config.json values
func ReadConfig() error {

	fmt.Println("Reading from config file...")

	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(file))

	err = json.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}

	BotPrefix = config.BotPrefix
	BotID = config.BotID
	ServerID = config.ServerID
	BotLogID = config.BotLogID
	CommandRoles = config.CommandRoles
	OptInUnder = config.OptInUnder
	OptInAbove = config.OptInAbove
	VoiceChaID = config.VoiceChaID
	Website = config.Website
	ModCategoryID = config.ModCategoryID
	VoteChannelCategoryID = config.VoteChannelCategoryID

	// Takes the bot token from the environment variable. Reason is to avoid pushing token to github
	if os.Getenv("ZeroTsuToken") == "" {
		panic("No token set in your environment variables for key \"ZeroTsuToken\"")
	}
	Token = os.Getenv("ZeroTsuToken")
	return nil
}

// Loads hidden configSecrets.json values
func ReadConfigSecrets() error {
	fmt.Println("Reading from configsecrets file...")

	file, err := ioutil.ReadFile("configsecrets.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(file, &configsecrets)
	if err != nil {
		panic(err)
	}

	RedditAppName = configsecrets.RedditAppName
	RedditAppSecret = configsecrets.RedditAppSecret
	DiscordAppSecret = configsecrets.DiscordAppSecret

	fmt.Println("Successfuly read configsecrets file.")

	return nil
}