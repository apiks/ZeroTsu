package commands

import 	"github.com/r-anime/ZeroTsu/config"

// The purpose of this init is to load the config values. The reason for this file is that init loads alphabetically so it has to be at the top
func init() {
	err := config.ReadConfig()
	if err != nil {
		panic(err)
	}
	err = config.ReadConfigSecrets()
	if err != nil {
		panic(err)
	}
}