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
	"time"

	_ "modernc.org/sqlite"
)

var GIT_COMMIT_HASH string

func main() {
	app := &App{
		GitCommitHash: GIT_COMMIT_HASH,
	}

	app.initialise()
	app.applicationLogFileOpen()
	app.loadHTMLTemplates()
	go app.monitorOperatingSystemSignals()

	app.DBAudioVaultOpen()
	app.DBAudioVaultGetSegments()

	// serve static files from the static sub-folder so that
	// they can be given appropriate Cache-Control HTTP Headers
	fileServer := http.FileServer(http.Dir(app.executableFolder + "static-assets/"))
	http.Handle("/static-assets/", app.webServerHeaders(app.webServerPassthrough(http.StripPrefix("/static-assets/", fileServer))))

	app.configureRoutes()

	fmt.Println("HTTP web service loaded.")
	fmt.Println("Press CTRL+C to exit & return to the terminal.")
	app.startWebServer()
	select {} // block, so the program stays resident
}

// set the default values for the application
func (app *App) initialise() {
	app.executableFolder = path.Dir(os.Args[0]) + "/"
	app.portNumber = ":1969"
	app.Testing = false

	hostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
	}

	switch hostname {
	case "signalsix":
		app.BaseURL = "http://localhost:1969/"
		app.Testing = true

		// on the developers machine create a GitCommitHash
		// based on the current date/time to cache burst any
		// changes made to templates, CSS & JS files
		now := time.Now()
		app.GitCommitHash = now.Format("2006-01-02-15-04-05")
	case "NUTH-VDS11":
		app.BaseURL = "https://audio-vault-uat.xnuth.nhs.uk/"
		app.Testing = true
	}

	app.GitCommitHashShort = app.GitCommitHash[0:8]
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
	app.DBAudioVaultClose()
	os.Exit(1)
}
