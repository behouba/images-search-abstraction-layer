package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "dbname=images_search_db user=postgres port=5000 sslmode=disable")
	if err != nil {
		fmt.Println("failed to connect to the database..", err)
		panic(err)
	}
}

func saveSearch(term string) error {
	t := time.Now()
	statement := "INSERT INTO IMAGES_RECORD (TERMS, DATES) VALUES ($1, $2)"
	_, err := db.Exec(statement, term, t.Format("Mon Jan 2 15:04:05"))
	if err != nil {
		return err
	}
	return nil
}

func getLatestSearch() ([]Search, error) {

	statement := "SELECT TERMS, DATES FROM IMAGES_RECORD ORDER BY ID DESC"
	rows, err := db.Query(statement)
	if err != nil {
		return []Search{}, err
	}
	var searchList []Search
	for rows.Next() {
		search := Search{}
		if err := rows.Scan(&search.Term, &search.When); err != nil {
			return []Search{}, err
		}
		searchList = append(searchList, search)
	}
	return searchList, nil
}
