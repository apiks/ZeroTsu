package config

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func EstablishConnection() {
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@(localhost)/%v", DBUser, DBPass, DBName))
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(1 * time.Minute)
	fmt.Println(db)

	stmt, err := db.Prepare("CREATE TABLE test(testValue varchar)")
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}
	stmt2, err := db.Prepare("INSERT INTO test(testValue) VALUES(?)")
	if err != nil {
		log.Fatal(err)
	}
	res2, err := stmt2.Exec("Dolly")
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res2.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res2.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)

	defer db.Close()
}