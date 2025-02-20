package main

import (
	"fmt"
	html "html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
)

func main() {
	app := &App{}
	app.initialise()
	app.applicationLogFileOpen()
	app.loadHTMLTemplates()
	go app.monitorOperatingSystemSignals()

	// serve static files from the static sub-folder so that
	// they can be given appropriate Cache-Control HTTP Headers
	fileServer := http.FileServer(http.Dir(app.executableFolder + "static-assets/"))
	http.Handle("/static-assets/", app.webServerHeaders(app.webServerPassthrough(http.StripPrefix("/static-assets/", fileServer))))

	app.configureRoutes()

	fmt.Println("HTTP web service loaded.")
	fmt.Println("Press return to return to terminal.")
	app.startWebServer()
	select {} // block, so the program stays resident
}

// set the default values for the application
func (app *App) initialise() {
	app.executableFolder = path.Dir(os.Args[0]) + "/"
	app.portNumber = ":1969"

	hostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
	}

	switch hostname {
	case "signalsix":
		app.BaseURL = "http://localhost:1969/"
	case "NUTH-VDS11":
		app.BaseURL = "https://audio-vault-uat.xnuth.nhs.uk/"
	}

	app.signalChannel = make(chan os.Signal, 1)
	signal.Notify(app.signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
}

func (app *App) loadHTMLTemplates() {
	app.tplHTML = html.Must(html.ParseGlob(path.Dir(os.Args[0]) + "/views/*.html"))
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
	app.applicationLogFileClose()
	os.Exit(1)
}
