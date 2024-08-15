package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

func runQuery(query string) *sql.Rows {
	var (
		dbPassword, _   = os.LookupEnv("DB_PASSWORD")
		dbUser, _       = os.LookupEnv("DB_USER")
		dbName, _       = os.LookupEnv("DB_NAME")
		dbDriverName, _ = os.LookupEnv("DB_DRIVER_NAME")
	)
	connStr := fmt.Sprintf("user=%v password=%v dbname=%v sslmode=disable", dbUser, dbPassword, dbName)
	db, err := sql.Open(dbDriverName, connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	result, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	return result
}

func setHash(guid string, refresh string) {
	var query = fmt.Sprintf("update \"Users\" set refresh = '%v' where guid = '%v'", refresh, guid)
	result := runQuery(query)
	defer result.Close()
}

func getHash(guid interface{}) string {
	var query = fmt.Sprintf("select refresh from \"Users\" where guid = '%v'", guid)

	result := runQuery(query)
	defer result.Close()

	var refresh string
	for result.Next() {
		err := result.Scan(&refresh)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	return refresh
}

func checkGuid(guid string) bool {
	var query = fmt.Sprintf("select count(*) from \"Users\" where guid = '%v'", guid)

	result := runQuery(query)
	defer result.Close()

	var count string
	for result.Next() {
		err := result.Scan(&count)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	if count == "0" {
		return false
	} else {
		return true
	}
}
