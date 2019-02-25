package config

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/go-sql-driver/mysql"
)

func EstablishConnection() {
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@(localhost)/%v", DBUser, DBPass, DBName))
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
}