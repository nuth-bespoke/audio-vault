package main

import (
	"database/sql"
	"log"
	"os"
	"runtime"

	_ "modernc.org/sqlite"
)

func (app *App) DBAudioVaultClose() {
	app.sqliteReader.Close()
}

func (app *App) DBAudioVaultGetSegments() string {
	var err error
	var rows *sql.Rows
	var html string

	// TODO: Convert inline html to template parse
	// TODO: Return all fields

	rows, err = app.sqliteReader.Query("SELECT DocumentID FROM Segments;")
	if err != nil {
		log.Println(err.Error())
	}

	html = html + `<ul>`
	for rows.Next() {
		var i string
		if err = rows.Scan(&i); err != nil {
			log.Println(err.Error())
		}
		html = html + `<li>` + i + `</li>`
	}
	html = html + `</ul>`

	return html
}

func (app *App) DBAudioVaultOpen() {
	var err error

	app.sqliteReader, err = sql.Open("sqlite", app.executableFolder+"audio-vault.db")
	if err != nil {
		log.Println("FATAL:Opening SQLite Reader :" + err.Error())
		os.Exit(1)
	}

	app.sqliteWriter, err = sql.Open("sqlite", app.executableFolder+"audio-vault.db")
	if err != nil {
		log.Println("FATAL:Opening SQLite Writer :" + err.Error())
		os.Exit(1)
	}

	app.sqliteWriter.SetMaxOpenConns(1)
	app.sqliteReader.SetMaxOpenConns(max(4, runtime.NumCPU()))
	app.DBAudioVaultSetPragmas()
}

func (app *App) DBAudioVaultSetPragmas() {
	var pragmas = `
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 1000000000;
PRAGMA foreign_keys = true;
PRAGMA temp_store = memory;`

	_, err := app.sqliteReader.Exec(pragmas)
	if err != nil {
		log.Println("FATAL:Setting Pragmas on SQLite Reader :" + err.Error())
		os.Exit(1)
	}

	_, err = app.sqliteWriter.Exec(pragmas)
	if err != nil {
		log.Println("FATAL:Setting Pragmas on SQLite Writer :" + err.Error())
		os.Exit(1)
	}
}
