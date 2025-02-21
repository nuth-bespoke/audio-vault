package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func (app *App) DBAudioVaultOpen() {
	var err error

	app.sqlite, err = sql.Open("sqlite", app.executableFolder+"audio-vault.db")
	if err != nil {
		fmt.Println(err.Error())
		log.Println(err.Error())
		os.Exit(1)
	}
}

func (app *App) DBAudioVaultClose() {
	app.sqlite.Close()
}

func (app *App) DBAudioVaultGetSegments() string {
	var err error
	var rows *sql.Rows
	var html string

	// TODO: Convert inline html to template parse
	// TODO: Return all fields

	rows, err = app.sqlite.Query("SELECT DocumentID FROM Segments;")
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
