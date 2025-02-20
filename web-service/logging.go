package main

import (
	"log"
	"os"
)

// Opens the applications log file to allow new entries to be appended to the log
// The web server should exit if the access file can't be created/opened.
func (app *App) applicationLogFileOpen() {
	var err error
	app.applicationLogFile, err = os.OpenFile(
		app.executableFolder+"audio-vault.log",
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
