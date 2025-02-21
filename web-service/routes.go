package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

func (app *App) routeHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte("AudioVault: OK :-)"))
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
