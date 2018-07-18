package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

//File for Bot, server, channel and role info.

var (
	Token        string
	BotPrefix    string
	BotID        string
	ServerID     string
	BotLogID     string
	CommandRoles []string
	OptInUnder   string
	OptInAbove   string

	config *configStruct
)

type configStruct struct {
	Token        string   `json:"-"`
	BotPrefix    string   `json:"BotPrefix"`
	BotID        string   `json:"BotID"`
	ServerID     string   `json:"ServerID"`
	BotLogID     string   `json:"BotLogID"`
	CommandRoles []string `json:"CommandRoles"`
	OptInUnder   string   `json:"OptInUnder"`
	OptInAbove   string   `json:"OptInAbove"`
}

func ReadConfig() error {

	fmt.Println("Reading from config file...")

	file, err := ioutil.ReadFile("config.json")
	if err != nil {

		fmt.Println(err.Error())
		return err
	}

	fmt.Println(string(file))

	err = json.Unmarshal(file, &config)
	if err != nil {

		fmt.Println(err.Error())
		return err
	}

	BotPrefix = config.BotPrefix
	BotID = config.BotID
	ServerID = config.ServerID
	BotLogID = config.BotLogID
	CommandRoles = config.CommandRoles
	OptInUnder = config.OptInUnder
	OptInAbove = config.OptInAbove

	// Takes the bot token from the environment variable. Reason is to avoid pushing token to github
	if os.Getenv("ZeroTsuToken") == "" {
		panic("No token set in your environment variables for key \"ZeroTsuToken\"")
	}
	Token = os.Getenv("ZeroTsuToken")
	return nil
}
