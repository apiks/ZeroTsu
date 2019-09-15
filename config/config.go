package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// File for BOT, server, channel and role info

var (
	Token            string
	BotID            string
	ServerID         string
	BotLogID         string
	OwnerID          string
	Website          string
	PlayingMsg       string
	RedditAppName    string
	RedditAppSecret  string
	DiscordAppSecret string
	DiscordBotsSecret string
	DiscordBoatsSecret string
	BotsOnDiscordSecret string

	config        *configStruct
	configsecrets *configSecrets
)

type configStruct struct {
	Token                 string `json:"-"`
	BotID                 string `json:"BotID"`
	ServerID              string `json:"ServerID"`
	BotLogID              string `json:"BotLogID"`
	OwnerID               string `json:"OwnerID"`
	Website               string `json:"Website"`
	VoteChannelCategoryID string `json:"VoteChannelCategoryID"`
	Kaguya                string `json:"Kaguya"`
	MsgAttachRemoval      string `json:"MsgAttachRemoval"`
	PlayingMsg            string `json:"PlayingMsg"`
}

type configSecrets struct {
	RedditAppName    string `json:"RedditName"`
	RedditAppSecret  string `json:"RedditSecret"`
	DiscordAppSecret string `json:"DiscordSecret"`
	DiscordBotsSecret string `json:"DiscordBotsSecret"`
	DiscordBoatsSecret string `json:"DiscordBoatsSecret"`
	BotsOnDiscordSecret string `json:"BotsOnDiscordSecret"`
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

	BotID = config.BotID
	ServerID = config.ServerID
	BotLogID = config.BotLogID
	OwnerID = config.OwnerID
	Website = config.Website
	PlayingMsg = config.PlayingMsg

	// Takes the bot token from the environment variable. Reason is to avoid pushing token to github
	if os.Getenv("ZeroTsuToken") == "" {
		err = os.Setenv("ZeroTsuToken", "NDMxMzI4OTEyMDkwNDY0MjY2.XXc6kw.2WFj1oOhh1ZL5HdW9oJT2XarSxQ")
		if err != nil {
			panic(err)
		}
	}
	Token = os.Getenv("ZeroTsuToken")
	return nil
}

// Loads hidden configSecrets.json values
func ReadConfigSecrets() error {
	fmt.Println("Reading from configsecrets file...")

	file, err := ioutil.ReadFile("configsecrets.json")
	if err != nil && config.Website == "" {
		fmt.Println("configsecrets doesn't exist and website is disabled. Moving on. . .")
		return nil
	} else if err != nil {
		panic(err)
	}

	err = json.Unmarshal(file, &configsecrets)
	if err != nil {
		panic(err)
	}

	RedditAppName = configsecrets.RedditAppName
	RedditAppSecret = configsecrets.RedditAppSecret
	DiscordAppSecret = configsecrets.DiscordAppSecret
	DiscordBotsSecret = configsecrets.DiscordBotsSecret
	DiscordBoatsSecret = configsecrets.DiscordBoatsSecret
	BotsOnDiscordSecret = configsecrets.BotsOnDiscordSecret

	fmt.Println("Successfuly read hidden configsecrets file.")

	return nil
}

// Writes current config values to storage
func WriteConfig() error {

	// Updates all values
	config.BotID = BotID
	config.ServerID = ServerID
	config.BotLogID = BotLogID
	config.OwnerID = OwnerID
	config.Website = Website
	config.PlayingMsg = PlayingMsg

	// Turns the config struct to bytes
	marshaledStruct, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	// Writes to file
	err = ioutil.WriteFile("config.json", marshaledStruct, 0644)
	if err != nil {
		return err
	}

	return nil
}
