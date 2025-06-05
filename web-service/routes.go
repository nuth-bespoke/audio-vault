package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

func (app *App) routeDictation(w http.ResponseWriter, r *http.Request) {
	var err error
	var tplBuffer bytes.Buffer
	var documentID string

	s := requestState{Application: *app}
	s.Dictation.WaveformExists = false
	s.Dictation.DictationAudioExists = false

	if strings.ToUpper(r.Method) == "GET" {
		documentID = strings.Replace(r.URL.Path, "/dictation/", "", -1)
		if _, err := strconv.Atoi(documentID); err == nil {
			s.Dictation.DocumentID = documentID
		} else {
			s.Dictation.DocumentID = "0"
		}

		s.WebPageTitle = "Dictation (" + s.Dictation.DocumentID + ")"

		if app.checkFileExists(app.executableFolder + "vault/dictations/" + s.Dictation.DocumentID + ".png") {
			s.Dictation.WaveformExists = true
		}
		if app.checkFileExists(app.executableFolder + "vault/dictations/" + s.Dictation.DocumentID + ".wav") {
			s.Dictation.DictationAudioExists = true
		}

		s.Dictation.SegmentHTML = app.DBAudioVaultGetSegmentsDataByDocumentID(s.Dictation.DocumentID)

		auditEventsIDs := app.DBAudioVaultGetSegmentsByDocumentID(s.Dictation.DocumentID)
		auditEventsIDs = append(auditEventsIDs, s.Dictation.DocumentID)
		auditEventsIDs = append(auditEventsIDs, s.Dictation.DocumentID+".wav")

		// build an SQL IN statement to grab all audit events
		instr := "[" + strings.Join(auditEventsIDs, "^") + "]"
		instr = strings.ReplaceAll(instr, `^`, `', '`)
		instr = strings.ReplaceAll(instr, `[`, `IN ('`)
		instr = strings.ReplaceAll(instr, `]`, `')`)

		s.Dictation.AuditEventsHTML = app.DBAudioVaultGetAudioEvents(instr)

		err = app.tplHTML.ExecuteTemplate(&tplBuffer, "dictation", s)
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

func (app *App) routeOrphan(w http.ResponseWriter, r *http.Request) {
	var err error
	var tplBuffer bytes.Buffer
	var MRN string

	s := requestState{Application: *app}

	if strings.ToUpper(r.Method) == "GET" {
		MRN = strings.Replace(r.URL.Path, "/mrn/", "", -1)
		if len(MRN) == 0 {
			MRN = "0"
		}
		s.Orphan.MRN = MRN

		s.WebPageTitle = "Orphan(s) (" + s.Orphan.MRN + ")"
		s.Orphan.OrphansHTML = template.HTML(app.DBAudioVaultGetOrphans(true, s.Orphan.MRN))

		err = app.tplHTML.ExecuteTemplate(&tplBuffer, "orphan", s)
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
	app.DBAudioVaultInsertAuditEvent(audioFilename, "stream [user played the audio for "+audioFilename+"]")

	switch r.Method {
	case http.MethodGet:

		var audioFolder string
		audioFolder = ""

		if app.checkFileExists("vault/segments/" + audioFilename) {
			audioFolder = "vault/segments/"
		}
		if app.checkFileExists("vault/dictations/" + audioFilename) {
			audioFolder = "vault/dictations/"
		}
		if app.checkFileExists("vault/orphans/" + audioFilename) {
			audioFolder = "vault/orphans/"
		}

		file, err = os.Open(app.executableFolder + audioFolder + audioFilename)
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

	uploadedOrphans := time.NewTicker(time.Second * 2)
	defer uploadedOrphans.Stop()

	completedDictations := time.NewTicker(time.Second * 2)
	defer completedDictations.Stop()

	pendingSegments := time.NewTicker(time.Second * 2)
	defer pendingSegments.Stop()

	cpuT := time.NewTicker(time.Second * 1)
	defer cpuT.Stop()

	clientGone := r.Context().Done()
	rc := http.NewResponseController(w)

	for {
		select {
		case <-clientGone:

		case <-uploadedOrphans.C:
			orphans := app.DBAudioVaultGetOrphans(false, "")

			//remove new lines from segments HTML so that
			//it can be sent over Server Side Events
			orphans = strings.Replace(orphans, "\n", "", -1)
			_, err := w.Write([]byte("event:orphans\ndata: " + orphans + "\n\n"))
			if err != nil {
				log.Println(err.Error())
				return
			}

		case <-completedDictations.C:
			dictations := app.DBAudioVaultGetDictations()

			//remove new lines from segments HTML so that
			//it can be sent over Server Side Events
			dictations = strings.Replace(dictations, "\n", "", -1)
			_, err := w.Write([]byte("event:dictations\ndata: " + dictations + "\n\n"))
			if err != nil {
				log.Println(err.Error())
				return
			}

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

func (app *App) routeOrphans(w http.ResponseWriter, r *http.Request) {
	var err error
	var file multipart.File
	var header *multipart.FileHeader
	var dst *os.File

	// if r.Header.Get("authorization") != "cf83e1357eefb8bdf1542850d66d800" {
	// 	log.Println("ERR: 401 Unauthorized, from " + r.RemoteAddr)
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	w.Write([]byte("401 Unauthorized"))
	// 	return
	// }

	if strings.ToUpper(r.Method) != "POST" {
		log.Println("ERR: 405 Method Not Allowed, from " + r.RemoteAddr)
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 Method Not Allowed"))
		return
	}

	// 1048576 = 1MB * 60 = Max Size is 60MB
	// A 2.19 minute diction in testing was 52MB
	err = r.ParseMultipartForm(1048576 * 60)
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

	err = r.ParseForm()
	if err != nil {
		log.Println("ERR: 400 Bad Request " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request" + err.Error()))
		return
	}

	submission := submission{}
	submission.MRN = r.Form.Get("MRN")
	submission.CreatedBy = strings.ToUpper(r.Form.Get("CreatedBy"))
	submission.MachineName = strings.ToUpper(r.Form.Get("MachineName"))
	submission.SegmentFileName = header.Filename
	submission.SegmentFileSize = strconv.FormatInt(header.Size, 10)

	// if app.Testing {
	// 	fmt.Println("MRN=" + submission.MRN)
	// 	fmt.Println("CreatedBy=" + submission.CreatedBy)
	// 	fmt.Println("MachineName=" + submission.MachineName)
	// 	fmt.Println(header.Filename)
	// 	fmt.Println("--------------------------")
	// }

	// create the file on disk in the vault folders
	dst, err = os.Create(app.executableFolder + "vault/orphans/" + header.Filename)
	if err != nil {
		errorMessage := "ERR: 500 error creating file entry for orphan " + header.Filename + " : " + err.Error()
		log.Println(errorMessage)
		app.DBAudioVaultInsertAuditEvent("0404", errorMessage)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}
	defer dst.Close()

	// copy the file to the new file location
	_, err = io.Copy(dst, file)
	if err != nil {
		errorMessage := "ERR: 500 error writing data to orphan file " + header.Filename + " : " + err.Error()
		log.Println(errorMessage)
		app.DBAudioVaultInsertAuditEvent("0404", errorMessage)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}

	app.DBAudioVaultInsertOrphan(&submission)
	app.DBAudioVaultInsertAuditEvent("0404", "docstore audio submission ["+header.Filename+" : "+strconv.FormatInt(header.Size, 10)+" bytes written]")

	log.Println("201 - " + header.Filename + " Created")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("201 - " + header.Filename + " Created"))
}

func (app *App) routeStore(w http.ResponseWriter, r *http.Request) {
	var err error
	var file multipart.File
	var header *multipart.FileHeader
	var dst *os.File

	// if r.Header.Get("authorization") != "cf83e1357eefb8bdf1542850d66d800" {
	// 	log.Println("ERR: 401 Unauthorized, from " + r.RemoteAddr)
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	w.Write([]byte("401 Unauthorized"))
	// 	return
	// }

	if strings.ToUpper(r.Method) != "POST" {
		log.Println("ERR: 405 Method Not Allowed, from " + r.RemoteAddr)
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 Method Not Allowed"))
		return
	}

	// 1048576 = 1MB * 60 = Max Size is 60MB
	// A 2.19 minute diction in testing was 52MB
	err = r.ParseMultipartForm(1048576 * 60)
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

	err = r.ParseForm()
	if err != nil {
		log.Println("ERR: 400 Bad Request " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request" + err.Error()))
		return
	}

	submission := submission{}
	submission.DocumentID = r.Form.Get("DocumentID")
	submission.MRN = r.Form.Get("MRN")
	submission.CreatedBy = strings.ToUpper(r.Form.Get("CreatedBy"))
	submission.MachineName = strings.ToUpper(r.Form.Get("MachineName"))
	submission.SegmentCount = r.Form.Get("SegmentCount")
	submission.SegmentOrder = r.Form.Get("SegmentOrder")
	submission.SegmentFileName = header.Filename
	submission.SegmentFileSize = strconv.FormatInt(header.Size, 10)

	// if app.Testing {
	// 	fmt.Println("DocID=" + submission.DocumentID)
	// 	fmt.Println("MRN=" + submission.MRN)
	// 	fmt.Println("CreatedBy=" + submission.CreatedBy)
	// 	fmt.Println("MachineName=" + submission.MachineName)
	// 	fmt.Println("SegmentCount=" + submission.SegmentCount)
	// 	fmt.Println("SegmentOrder=" + submission.SegmentOrder)
	// 	fmt.Println(header.Filename)
	// 	fmt.Println("--------------------------")
	// }

	// create the file on disk in the vault folders
	dst, err = os.Create(app.executableFolder + "vault/segments/" + header.Filename)
	if err != nil {
		errorMessage := "ERR: 500 error creating file entry for " + header.Filename + " : " + err.Error()
		log.Println(errorMessage)
		app.DBAudioVaultInsertAuditEvent(submission.DocumentID, errorMessage)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}
	defer dst.Close()

	// copy the file to the new file location
	_, err = io.Copy(dst, file)
	if err != nil {
		errorMessage := "ERR: 500 error writing data to file " + header.Filename + " : " + err.Error()
		log.Println(errorMessage)
		app.DBAudioVaultInsertAuditEvent(submission.DocumentID, errorMessage)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}

	app.DBAudioVaultInsertDictation(&submission)
	app.DBAudioVaultInsertSegment(&submission)
	app.DBAudioVaultInsertAuditEvent(submission.DocumentID, "docstore audio submission ["+header.Filename+" : "+strconv.FormatInt(header.Size, 10)+" bytes written]")

	log.Println("201 - " + header.Filename + " Created")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("201 - " + header.Filename + " Created"))
}

func (app *App) routeDashboard(w http.ResponseWriter, r *http.Request) {
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

func GenerateUserMD5Hash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
}

func (app *App) routeUser(w http.ResponseWriter, r *http.Request) {
	var err error
	var tplBuffer bytes.Buffer

	s := requestState{Application: *app}

	if strings.ToUpper(r.Method) == "GET" {
		s.UserID = strings.Replace(r.URL.Path, "/user/", "", -1)
		s.UserID = strings.ToUpper(s.UserID)
		s.UserHash = GenerateUserMD5Hash(s.UserID)

		s.WebPageTitle = "User Logs (" + s.UserID + ")"

		err = app.tplHTML.ExecuteTemplate(&tplBuffer, "user", s)
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

func (app *App) routeWaveForm(w http.ResponseWriter, r *http.Request) {
	var waveFormFilename string
	var err error
	var file *os.File

	waveFormFilename = strings.Replace(r.URL.Path, "/waveform/", "", -1)

	switch r.Method {
	case http.MethodGet:

		file, err = os.Open(app.executableFolder + "vault/dictations/" + waveFormFilename)
		if err != nil {
			log.Println("ERR: opening /waveform/" + waveFormFilename)
		}

		defer func() {
			err := file.Close()
			if err != nil {
				log.Println("ERR: closing /waveform/" + waveFormFilename)
			}
		}()

		w.Header().Set("Cache-Control", "public, max-age=63072000, immutable")
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Server", "AuditVault")
		w.WriteHeader(http.StatusOK)

		_, err = io.Copy(w, file)
		if err != nil {
			log.Println("ERR: writing /waveform/" + waveFormFilename)
		}
	}
}
