package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

func (app *App) routeHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte("AudioVault: OK :-)"))
}

func (app *App) routeStream(w http.ResponseWriter, r *http.Request) {
	var audioFilename string
	var err error
	var file *os.File

	audioFilename = strings.Replace(r.URL.Path, "/stream/", "", -1)

	switch r.Method {
	case http.MethodGet:

		file, err = os.Open(app.executableFolder + "vault/segments/" + audioFilename)
		if err != nil {
			log.Println("ERR: opening /stream/" + audioFilename)
		}

		defer func() {
			err := file.Close()
			if err != nil {
				log.Println("ERR: closing /stream/" + audioFilename)
			}
		}()

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "audio/wav")

		_, err = io.Copy(w, file)
		if err != nil {
			log.Println("ERR: writing /stream/" + audioFilename)
		}
	}
}

func (app *App) routeServerSideEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	pendingSegments := time.NewTicker(time.Second * 2)
	defer pendingSegments.Stop()

	cpuT := time.NewTicker(time.Second * 1)
	defer cpuT.Stop()

	clientGone := r.Context().Done()
	rc := http.NewResponseController(w)

	for {
		select {
		case <-clientGone:

		case <-pendingSegments.C:
			segments := app.DBAudioVaultGetSegments()

			//remove new lines from segments HTML so that
			//it can be sent over Server Side Events
			segments = strings.Replace(segments, "\n", "", -1)
			_, err := w.Write([]byte("event:segments\ndata: " + segments + "\n\n"))
			if err != nil {
				log.Println(err.Error())
				return
			}

		case <-cpuT.C:
			c, err := cpu.Percent(0, true)
			if err != nil {
				log.Printf("unable to get cpu: %s", err.Error())
				return
			}

			_, err = fmt.Fprintf(w, "event:cpu\ndata: %.2f\n\n", c[0])
			if err != nil {
				log.Println(err.Error())
				return
			}

			rc.Flush()
		}
	}
}

func (app *App) routeStore(w http.ResponseWriter, r *http.Request) {
	var err error
	var file multipart.File
	var header *multipart.FileHeader

	if r.Header.Get("authorization") != "cf83e1357eefb8bdf1542850d66d800" {
		log.Println("ERR: 401 Unauthorized, from " + r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 Unauthorized"))
		return
	}

	if strings.ToUpper(r.Method) != "POST" {
		log.Println("ERR: 405 Method Not Allowed, from " + r.RemoteAddr)
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 Method Not Allowed"))
		return
	}

	err = r.ParseMultipartForm(1024 * 4)
	if err != nil {
		log.Println("ERR: 500 error parsing form data " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}

	file, header, err = r.FormFile("fileupload")
	if err != nil {
		log.Println("ERR: 500 error parsing form data " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}
	defer file.Close()

	// create the file on disk in the vault folders
	dst, err := os.Create(app.executableFolder + "vault/segments/" + header.Filename)
	if err != nil {
		log.Println("ERR: 500 error creating file entry " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}
	defer dst.Close()

	// copy the file to the new file location
	if _, err := io.Copy(dst, file); err != nil {
		log.Println("ERR: 500 error writing data to file " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}

	// w.Header().Set("K", "V")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 - " + header.Filename + " received"))
}

func (app *App) routeTesting(w http.ResponseWriter, r *http.Request) {
	var err error
	var tplBuffer bytes.Buffer

	s := requestState{Application: *app}

	if strings.ToUpper(r.Method) == "GET" {

		s.WebPageTitle = "Testers Dashboard"

		err = app.tplHTML.ExecuteTemplate(&tplBuffer, "dashboard", s)
		if err != nil {
			log.Println("ERR:" + err.Error())
		}

		w.WriteHeader(200)
		w.Write(tplBuffer.Bytes())
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed. :-("))
	}
}
