package main

import (
	"database/sql"
	"embed"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

//go:embed sql/*
var sqlFiles embed.FS

func init() {
	var err error
	db, err = sql.Open("sqlite3", "data.db")
	if err != nil {
		logError(err)
		return
	}

	buf, err := sqlFiles.ReadFile("sql/init.sql")
	if err != nil {
		logError(err)
		return
	}
	var body string = string(buf)

	if _, err := db.Exec(body); err != nil {
		logError(err)
		return
	}
}
