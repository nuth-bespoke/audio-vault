package main

import (
	html "html/template"
	"os"
)

type App struct {
	accessLogFile      *os.File       // file handle to web servers access logs
	applicationLogFile *os.File       // file handle to application logs
	executableFolder   string         // the folder the binary was executed from
	portNumber         string         // the port number to run the web server on
	signalChannel      chan os.Signal // channel to monitor operating system signals
	tplHTML            *html.Template // pointer to all the HTML templates (views)
}
