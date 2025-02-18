package main

import (
	"log"
	"os"
)

// Opens the web servers log file to allow new entries to be appended to the log
// The web server should exit if the access file can't be created/opened.
func (app *App) accessLogFileOpen() {
	var err error
	app.accessLogFile, err = os.OpenFile(
		app.executableFolder+"access.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666)
	if err != nil {
		println("ERROR:" + err.Error())
		os.Exit(1)
	}
}

// Flush any pending log entries to stable storage before closing the access log file.
func (app *App) accessLogFileClose() {
	app.accessLogFile.Sync()
	app.accessLogFile.Close()
}

// Opens the applications log file to allow new entries to be appended to the log
// The web server should exit if the access file can't be created/opened.
func (app *App) applicationLogFileOpen() {
	var err error
	app.applicationLogFile, err = os.OpenFile(
		app.executableFolder+"audit-vault.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666)
	if err != nil {
		println("ERROR:" + err.Error())
		os.Exit(1)
	}
	log.SetOutput(app.applicationLogFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// Flush any pending log entries to stable storage before closing the access log file.
func (app *App) applicationLogFileClose() {
	app.applicationLogFile.Sync()
	app.applicationLogFile.Close()
}
