package main

import (
	"errors"
	"fmt"
	html "html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
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
	app.createFolderStructure()
	app.loadHTMLTemplates()
	go app.monitorOperatingSystemSignals()

	app.DBAudioVaultOpen()
	app.DBAudioVaultGetSegments()

	// serve static files from the static sub-folder so that
	// they can be given appropriate Cache-Control HTTP Headers
	fileServer := http.FileServer(http.Dir(app.executableFolder + "static-assets/"))
	http.Handle("/static-assets/", app.webServerHeaders(app.webServerPassthrough(http.StripPrefix("/static-assets/", fileServer))))

	go app.SoxGetMetadata()
	go app.SoxNormaliseSegments()
	go app.SoxConcatenateSegments()

	app.configureRoutes()
	fmt.Println("HTTP web service loaded.")
	fmt.Println("Press CTRL+C to exit & return to the terminal.")
	app.startWebServer()
	select {} // block, so the program stays resident
}

func (app *App) createFolderTree(path string) {
	tree := app.executableFolder + path
	err := os.MkdirAll(tree, os.ModePerm)
	if err != nil {
		log.Println("FATAL:" + err.Error())
		fmt.Println("FATAL:" + err.Error())
		os.Exit(1)
	}
}

func (app *App) checkFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}

func (app *App) createFolderStructure() {
	app.createFolderTree("vault/dictations/")
	app.createFolderTree("vault/orphans/")
	app.createFolderTree("vault/segments/")
}

func (app *App) executeExternalCommand(executable string, arguments []string) (errorCode int, errorMessage string) {
	var err error
	var errorMsg string
	var commandContext string
	var cmd *exec.Cmd
	var out []byte

	errorMsg = ""
	commandContext = " :" + executable + " [" + strings.Join(arguments, " ") + "]"

	log.Println("LAUNCHING" + commandContext)

	cmd = exec.Command(executable, arguments...)
	out, err = cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMsg = "ERR" + commandContext
			log.Println(errorMsg)
			log.Println(string(out))
			return exitError.ExitCode(), errorMsg
		}
		errorMsg = "ERR: could not run [" + executable + "] " + err.Error()
		return -1, errorMsg
	}

	log.Println("SUCCESS" + commandContext)
	return 0, ""
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

	if runtime.GOOS == "windows" {
		app.soxExecutable = app.executableFolder + "tools/sox/sox.exe"
		app.audioWaveFormExecutable = app.executableFolder + "tools/audiowaveform/audiowaveform.exe"
	} else {
		app.soxExecutable = "/usr/bin/sox"
		app.audioWaveFormExecutable = "/usr/bin/audiowaveform"
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

func (app *App) SoxParseMetadata(prefix string, data []string) string {
	for _, row := range data {
		if strings.HasPrefix(row, prefix) {
			if prefix == "Duration" {
				// duration uses different parsing as its got the
				// time indicators which are also : characters
				values := strings.Split(row, ":")
				return strings.TrimSpace(values[1]) + strings.TrimSpace(values[2]) + strings.TrimSpace(values[3])
			} else {
				values := strings.Split(row, ":")
				return strings.TrimSpace(values[1])
			}
		}
	}
	return ""
}

func (app *App) SoxGetMetadata() {
	var timerEnabled bool

	timerEnabled = true
	for {
		if timerEnabled {
			timerEnabled = false

			segments := strings.Split(app.DBAudioVaultGetSegmentsByProgressID(0), `^`)
			for _, filename := range segments {
				if len(filename) == 0 {
					break
				}

				filenamePath := app.executableFolder + "vault/segments/" + filename
				if app.checkFileExists(filenamePath) {

					sox_args := []string{"--info", filenamePath}
					log.Println("INFO: sox ", sox_args)

					cmd := exec.Command(app.soxExecutable, sox_args...)
					out, err := cmd.Output()
					if err != nil {
						if exitError, ok := err.(*exec.ExitError); ok {
							errorMessage := "ERR: sox --info return code " + strconv.Itoa(exitError.ExitCode())
							log.Println(errorMessage)
							app.DBAudioVaultInsertAuditEvent(filename, errorMessage)
							app.DBAudioVaultUpdateSegmentSoxReturnCode(filename, exitError.ExitCode())
							break
						}
						log.Println("ERR: could not run command, ", err.Error())
						break
					}

					lines := strings.Split(string(out), "\n")
					app.DBAudioVaultUpdateSegmentMetadata(
						app.SoxParseMetadata("Bit Rate", lines),
						app.SoxParseMetadata("Duration", lines),
						app.SoxParseMetadata("Precision", lines),
						app.SoxParseMetadata("Sample Rate", lines),
						filename)

					app.DBAudioVaultInsertAuditEvent(filename, "sox []"+strings.Join(sox_args, " ")+"]")
					app.DBAudioVaultInsertAuditEvent(filename,
						"sox --info successful with "+
							strconv.Itoa(len(lines))+
							" meta data items returned")
				}
			}

			time.Sleep(5 * time.Second)
			timerEnabled = true
		}
	}
}

func (app *App) SoxNormaliseSegments() {
	var timerEnabled bool

	timerEnabled = true
	for {
		if timerEnabled {
			timerEnabled = false

			segments := strings.Split(app.DBAudioVaultGetSegmentsByProgressID(1), `^`)
			for _, filename := range segments {
				if len(filename) == 0 {
					break
				}

				filenamePath := app.executableFolder + "vault/segments/" + filename
				if app.checkFileExists(filenamePath) {

					sox_args := []string{"--norm", filenamePath, "-r 48000", "-c 1", filenamePath + ".normal.wav"}
					log.Println("INFO: sox ", sox_args)

					cmd := exec.Command(app.soxExecutable, sox_args...)
					out, err := cmd.Output()
					if err != nil {
						if exitError, ok := err.(*exec.ExitError); ok {
							errorMessage := "ERR: sox --norm return code " + strconv.Itoa(exitError.ExitCode())
							log.Println(errorMessage)
							app.DBAudioVaultInsertAuditEvent(filename, errorMessage)
							app.DBAudioVaultUpdateSegmentSoxReturnCode(filename, exitError.ExitCode())
							break
						}
						log.Println("ERR: could not run command, ", err.Error())
						break
					}

					_ = out
					app.DBAudioVaultUpdateSegmentNormalised(filename)
					app.DBAudioVaultInsertAuditEvent(filename, "sox ["+strings.Join(sox_args, " ")+"]")
				}
			}

			time.Sleep(5 * time.Second)
			timerEnabled = true
		}
	}
}

func (app *App) SoxConcatenateSegments() {
	var timerEnabled bool

	timerEnabled = true
	for {
		if timerEnabled {
			timerEnabled = false

			DocumentIDs := app.DBAudioVaultGetSegmentsReadyForConcatConcatenation()
			if len(DocumentIDs) > 0 {
				for _, documentID := range DocumentIDs {

					sox_args := []string{"--combine", "concatenate"}
					filenames := app.DBAudioVaultGetSegmentsByDocumentID(documentID)
					if len(filenames) > 0 {
						for _, filename := range filenames {
							if len(filename) >= 1 {
								sox_args = append(sox_args, app.executableFolder+"vault/segments/"+filename+".normal.wav")
							}
						}
					}
					sox_args = append(sox_args, "vault/dictations/"+documentID+".wav")

					log.Println("INFO: sox ", sox_args)
					cmd := exec.Command(app.soxExecutable, sox_args...)

					out, err := cmd.CombinedOutput()
					if err != nil {
						if exitError, ok := err.(*exec.ExitError); ok {
							errorMessage := "ERR: sox --combine concatenate return code " + strconv.Itoa(exitError.ExitCode())
							log.Println(errorMessage)
							app.DBAudioVaultInsertAuditEvent(documentID, errorMessage)
							app.DBAudioVaultUpdateSegmentSoxReturnCode(documentID, exitError.ExitCode())
							break
						}
						log.Println("ERR: could not run command, ", err.Error())
						break
					}

					_ = out

					// generate a PNG of the audio wave form
					audio_wave_form_args := []string{
						"-i" + app.executableFolder + "vault/dictations/" + documentID + ".wav",
						"-o" + app.executableFolder + "vault/dictations/" + documentID + ".png",
						"-zauto", "-w800", "-h150",
					}
					errCode, errMessage := app.executeExternalCommand(app.audioWaveFormExecutable, audio_wave_form_args)
					if errCode != 0 {
						app.DBAudioVaultInsertAuditEvent(documentID, errMessage)
						break
					}

					app.DBAudioVaultInsertAuditEvent(documentID, "SUCCESS:"+app.audioWaveFormExecutable+" ["+strings.Join(audio_wave_form_args, " ")+"]")

					// TODO: create function to update Segments table
					// TODO: create function to update Dictations table
				}
			}

			time.Sleep(5 * time.Second)
			timerEnabled = true
		}
	}
}
