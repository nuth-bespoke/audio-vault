package main

import (
	"database/sql"
	html "html/template"
	"os"
)

type App struct {
	applicationLogFile *os.File       // file handle to application logs
	executableFolder   string         // the folder the binary was executed from
	portNumber         string         // the port number to run the web server on
	signalChannel      chan os.Signal // channel to monitor operating system signals
	tplHTML            *html.Template // pointer to all the HTML templates (views)
	sqliteReader       *sql.DB
	sqliteWriter       *sql.DB

	// Public variables which need to be accessed from the HTML templates/views
	BaseURL            string // the base URL of the application (different for different hosts)
	GitCommitHash      string // holds the latest Git Commit has from git rev-parse HEAD
	GitCommitHashShort string // holds the first 8 characters of the full Git Commit
	Testing            bool   // true if software is running on a host server that is a testing server
}

// requestState is the main data structure
// used by every request to the web service
type requestState struct {
	Application  App
	WebPageTitle string
}
