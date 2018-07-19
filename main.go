package main

import (
	"fmt"

	"github.com/r-anime/ZeroTsu/bot"
	"github.com/r-anime/ZeroTsu/config"
)

// Initializes and starts Bot and website
func main() {

	err := config.ReadConfig()

	if err != nil {

		fmt.Println(err.Error())
		return
	}

	bot.Start()

	// Web Server
	//http.HandleFunc("/", verification.IndexHandler)
	//http.Handle("/verification/", http.StripPrefix("/verification/", http.FileServer(http.Dir("verification"))))
	//err = http.ListenAndServe(":3000", nil)
	//if err != nil {
	//
	//	fmt.Println("Error:", err)
	//}

	<-make(chan struct{})
	return
}