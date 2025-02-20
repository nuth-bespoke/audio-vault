package main

import (
	html "html/template"
	"os"
)

type App struct {
	applicationLogFile *os.File       // file handle to application logs
	executableFolder   string         // the folder the binary was executed from
	portNumber         string         // the port number to run the web server on
	signalChannel      chan os.Signal // channel to monitor operating system signals
	tplHTML            *html.Template // pointer to all the HTML templates (views)
}

// requestState is the main data structure
// used by every request to the web service
type requestState struct {
	Application  App
	WebPageTitle string
}
