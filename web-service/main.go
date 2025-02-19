package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
)

func main() {
	app := &App{}
	app.initialise()
	app.accessLogFileOpen()
	app.applicationLogFileOpen()
	app.loadHTMLTemplates()
	go app.monitorOperatingSystemSignals()

	// serve static files from the static sub-folder so that
	// they can be given appropriate Cache-Control HTTP Headers
	// fileServer := http.FileServer(http.Dir("./static/"))
	// http.Handle("/static/", app.webServerHeaders(app.staticAccessLogs(http.StripPrefix("/static", fileServer))))
	// log.Println("INFO:static hosting invoked")

	app.configureRoutes()
	app.startWebServer()
	fmt.Println("HTTP web service loaded.")
	fmt.Println("Press return to return to terminal.")
	select {} // block, so the program stays resident
}

// set the default values for the application
func (app *App) initialise() {
	app.executableFolder = path.Dir(os.Args[0]) + "/"
	app.portNumber = ":1969"

	app.signalChannel = make(chan os.Signal, 1)
	signal.Notify(app.signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
}

func (app *App) loadHTMLTemplates() {
	// app.tplHTML = html.Must(html.ParseGlob(path.Dir(os.Args[0]) + "views/*.html"))
	log.Println("INFO:HTML Templates Loaded")
}

// monitors operating system signals and handle any event
// or interrupt logging the result to the master log file
func (app *App) monitorOperatingSystemSignals() {
	signalType := <-app.signalChannel
	signal.Stop(app.signalChannel)
	fmt.Println("")
	fmt.Println("***")
	fmt.Println("EXIT command received. Exiting...")
	switch signalType {
	case os.Interrupt:
		log.Println("FATAL: CTRL+C pressed")
	case syscall.SIGTERM:
		log.Println("FATAL: SIGTERM detected")
	}

	// flush log files and close them
	app.accessLogFileClose()
	app.applicationLogFileClose()
	os.Exit(1)
}
