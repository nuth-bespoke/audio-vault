package main

import (
	"bytes"
	"database/sql"
	"html/template"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/dustin/go-humanize"
	_ "modernc.org/sqlite"
)

func (app *App) DBAudioVaultClose() {
	app.sqliteReader.Close()
}

func (app *App) DBAudioVaultGetAudioEvents(instr string) template.HTML {
	var err error
	var rows *sql.Rows
	var tplBuffer bytes.Buffer

	auditEvents := auditEvents{}

	rows, err = app.sqliteReader.Query(`
		SELECT
			EventAt,
			EventMessage
		FROM AuditEvents
		WHERE SegmentFileName ` + instr + `
		ORDER BY EventAt ASC;`)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	for rows.Next() {
		var audioEvent auditEvent
		if err = rows.Scan(
			&audioEvent.EventAt,
			&audioEvent.EventMessage); err != nil {
			log.Println("ERR:" + err.Error())
		}
		auditEvents.AuditEvents = append(auditEvents.AuditEvents, audioEvent)
	}
	rows.Close()

	err = app.tplHTML.ExecuteTemplate(&tplBuffer, "audit-events", auditEvents)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	return (template.HTML(tplBuffer.String()))
}

func (app *App) DBAudioVaultGetDictationsForDocstore() docstoreDictations {
	var err error
	var rows *sql.Rows

	var docstoreDictationRows = docstoreDictations{}

	rows, err = app.sqliteReader.Query(`
		SELECT
			DocumentID,
			strftime('%Y-%d-%m %H:%M:%S', SavedAt) AS CreationDate,
			strftime('%Y-%m-%d %H:%M:%S', CompletedAt) AS DictationDate
		FROM Dictations
		WHERE CompletedAt IS NOT NULL
		  AND SentToDocstore IS NULL
		LIMIT 0, 10;`)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	for rows.Next() {
		var docstoreDictationRow docstoreDictation

		if err = rows.Scan(
			&docstoreDictationRow.DocumentID,
			&docstoreDictationRow.SavedAt,
			&docstoreDictationRow.DictatedAt); err != nil {
			log.Println("ERR:" + err.Error())
		}

		docstoreDictationRows.Dictations = append(docstoreDictationRows.Dictations, docstoreDictationRow)
	}

	rows.Close()
	return docstoreDictationRows
}

func (app *App) DBAudioVaultGetDictations() string {
	var err error
	var rows *sql.Rows
	var tplBuffer bytes.Buffer

	dictations := docstoreDictations{}

	rows, err = app.sqliteReader.Query(`
		SELECT
			DocumentID,
			MRN,
			CreatedBy,
			MachineName,
			IFNULL(strftime('%d-%m-%Y %H:%M:%S', SavedAt), "-") AS SavedAt,
			IFNULL(strftime('%d-%m-%Y %H:%M:%S', SentToDocstore), "-") AS SentToDocstore
		 FROM Dictations
		WHERE CompletedAt IS NOT NULL
		ORDER BY CompletedAt DESC;`)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	// IFNULL(fax, 'Call:' || phone) fax

	for rows.Next() {
		var dictation docstoreDictation

		if err = rows.Scan(
			&dictation.DocumentID,
			&dictation.MRN,
			&dictation.CreatedBy,
			&dictation.MachineName,
			&dictation.SavedAt,
			&dictation.SentToDocstore); err != nil {
			log.Println("ERR:" + err.Error())
		}

		dictations.Dictations = append(dictations.Dictations, dictation)
	}

	rows.Close()

	err = app.tplHTML.ExecuteTemplate(&tplBuffer, "dictations-listing", dictations)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	return tplBuffer.String()
}

func (app *App) DBAudioVaultGetSegments() string {
	var err error
	var rows *sql.Rows
	var tplBuffer bytes.Buffer

	segments := segments{}

	rows, err = app.sqliteReader.Query(`
		SELECT
			Segments.DocumentID,
			Dictations.CreatedBy,
			Dictations.MachineName,
			Segments.SegmentFileName,
			Segments.SegmentFileSize,
			Segments.AudioBitRate,
			Segments.AudioDuration,
			Segments.AudioPrecision,
			Segments.AudioSampleRate,
			Segments.ProcessingProgress,
			Segments.SoxStatusCode
		FROM Segments
		LEFT JOIN Dictations ON Segments.DocumentID = Dictations.DocumentID
		WHERE ProcessingProgress <= 2
		ORDER BY Segments.DocumentID, SegmentFileOrder;`)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	for rows.Next() {
		var audioSegment segment

		if err = rows.Scan(
			&audioSegment.DocumentID,
			&audioSegment.CreatedBy,
			&audioSegment.MachineName,
			&audioSegment.SegmentFileName,
			&audioSegment.SegmentFileSize,
			&audioSegment.AudioBitRate,
			&audioSegment.AudioDuration,
			&audioSegment.AudioPrecision,
			&audioSegment.AudioSampleRate,
			&audioSegment.ProcessingProgress,
			&audioSegment.SoxStatusCode); err != nil {
			log.Println("ERR:" + err.Error())
		}

		fileSize, _ := strconv.ParseUint(audioSegment.SegmentFileSize, 0, 64)
		audioSegment.SegmentFileSize = humanize.Bytes(fileSize)
		audioSegment.IncludePlayerControls = false
		segments.Segments = append(segments.Segments, audioSegment)
	}

	rows.Close()

	err = app.tplHTML.ExecuteTemplate(&tplBuffer, "segments-listing", segments)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	return tplBuffer.String()
}

func (app *App) DBAudioVaultGetSegmentsDataByDocumentID(id string) template.HTML {
	var err error
	var rows *sql.Rows
	var tplBuffer bytes.Buffer

	segments := segments{}

	rows, err = app.sqliteReader.Query(`
		SELECT
			Segments.DocumentID,
			Dictations.CreatedBy,
			Dictations.MachineName,
			Segments.SegmentFileName,
			Segments.SegmentFileSize,
			Segments.AudioBitRate,
			Segments.AudioDuration,
			Segments.AudioPrecision,
			Segments.AudioSampleRate,
			Segments.ProcessingProgress,
			Segments.SoxStatusCode
		FROM Segments
		LEFT JOIN Dictations ON Segments.DocumentID = Dictations.DocumentID
		WHERE Segments.DocumentID = ?
		ORDER BY Segments.DocumentID, SegmentFileOrder;`, id)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	for rows.Next() {
		var audioSegment segment

		if err = rows.Scan(
			&audioSegment.DocumentID,
			&audioSegment.CreatedBy,
			&audioSegment.MachineName,
			&audioSegment.SegmentFileName,
			&audioSegment.SegmentFileSize,
			&audioSegment.AudioBitRate,
			&audioSegment.AudioDuration,
			&audioSegment.AudioPrecision,
			&audioSegment.AudioSampleRate,
			&audioSegment.ProcessingProgress,
			&audioSegment.SoxStatusCode); err != nil {
			log.Println("ERR:" + err.Error())
		}

		fileSize, _ := strconv.ParseUint(audioSegment.SegmentFileSize, 0, 64)
		audioSegment.SegmentFileSize = humanize.Bytes(fileSize)
		audioSegment.IncludePlayerControls = true

		segments.Segments = append(segments.Segments, audioSegment)
	}

	rows.Close()

	err = app.tplHTML.ExecuteTemplate(&tplBuffer, "segments-listing", segments)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	return (template.HTML(tplBuffer.String()))
}

func (app *App) DBAudioVaultGetSegmentsByDocumentID(id string) []string {
	var err error
	var rows *sql.Rows
	var results []string

	rows, err = app.sqliteReader.Query(`
		SELECT
			SegmentFileName
		FROM Segments
		WHERE DocumentID = ?
		ORDER BY SegmentFileOrder`, id)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	for rows.Next() {
		var SegmentFileName string

		if err = rows.Scan(&SegmentFileName); err != nil {
			log.Println("ERR:" + err.Error())
		}

		results = append(results, SegmentFileName)
	}

	rows.Close()
	return results
}

func (app *App) DBAudioVaultGetSegmentsByProgressID(progressID int) string {
	var err error
	var rows *sql.Rows

	rows, err = app.sqliteReader.Query(`
		SELECT
			SegmentFileName
		 FROM Segments
		WHERE ProcessingProgress = ?
		ORDER BY DocumentID
		LIMIT 0, 10;`, progressID)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	var SegmentsToProcess string
	for rows.Next() {
		var SegmentFileName string

		if err = rows.Scan(&SegmentFileName); err != nil {
			log.Println("ERR:" + err.Error())
		}

		SegmentsToProcess = SegmentsToProcess + SegmentFileName + `^`
	}

	rows.Close()
	return SegmentsToProcess
}

func (app *App) DBAudioVaultGetSegmentsReadyForConcatConcatenation() []string {
	var err error
	var rows *sql.Rows
	var results []string

	rows, err = app.sqliteReader.Query(`
		SELECT
			d.DocumentID,
			d.SegmentCount,
			COUNT(s.DocumentID) AS Actual_Segment_Count
		FROM Dictations d
		LEFT JOIN Segments s ON d.DocumentID = s.DocumentID
		WHERE d.CompletedAt IS NULL
		GROUP BY d.DocumentID
		HAVING COUNT(s.DocumentID) = d.SegmentCount AND ProcessingProgress = 2
		LIMIT 0, 10;`)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	for rows.Next() {
		var DocumentID string
		var DictationsSegmentCount int
		var ReadySegmentCount int

		if err = rows.Scan(
			&DocumentID,
			&DictationsSegmentCount,
			&ReadySegmentCount); err != nil {
			log.Println("ERR:" + err.Error())
		}

		if DictationsSegmentCount == ReadySegmentCount {
			results = append(results, DocumentID)
		}
	}

	rows.Close()
	return results
}

func (app *App) DBAudioVaultOpen() {
	var err error

	app.sqliteReader, err = sql.Open("sqlite", app.executableFolder+"audio-vault.db")
	if err != nil {
		log.Println("FATAL:Opening SQLite Reader :" + err.Error())
		os.Exit(1)
	}

	app.sqliteWriter, err = sql.Open("sqlite", app.executableFolder+"audio-vault.db")
	if err != nil {
		log.Println("FATAL:Opening SQLite Writer :" + err.Error())
		os.Exit(1)
	}

	app.sqliteWriter.SetMaxOpenConns(1)
	app.sqliteReader.SetMaxOpenConns(max(4, runtime.NumCPU()))
	app.DBAudioVaultSetPragmas()
}

func (app *App) DBAudioVaultSetPragmas() {
	var pragmas = `
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 1000000000;
PRAGMA foreign_keys = true;
PRAGMA temp_store = memory;`

	_, err := app.sqliteReader.Exec(pragmas)
	if err != nil {
		log.Println("FATAL:Setting Pragmas on SQLite Reader :" + err.Error())
		os.Exit(1)
	}

	_, err = app.sqliteWriter.Exec(pragmas)
	if err != nil {
		log.Println("FATAL:Setting Pragmas on SQLite Writer :" + err.Error())
		os.Exit(1)
	}
}

func (app *App) DBAudioVaultUpdateSegmentNormalised(filename string) {

	var sql = `
		UPDATE Segments SET
			ProcessingProgress = 2
		WHERE SegmentFileName = ?`

	_, err := app.sqliteWriter.Exec(sql, filename)
	if err != nil {
		log.Println("FATAL:Updating Segments Normalised :" + err.Error())
	}
}

func (app *App) DBAudioVaultUpdateDocstoreCompletedDate(documentID string) {
	var sql = `
		UPDATE Dictations SET
			SentToDocstore = datetime(current_timestamp, 'localtime')
		WHERE DocumentID = ?`

	_, err := app.sqliteWriter.Exec(sql, documentID)
	if err != nil {
		log.Println("FATAL:Updating Dictation Docstore Date :" + err.Error())
	}
}

func (app *App) DBAudioVaultUpdateDictationComplete(documentID string) {

	var sql = `
		UPDATE Segments SET
			ProcessingProgress = 3
		WHERE DocumentID = ?`

	_, err := app.sqliteWriter.Exec(sql, documentID)
	if err != nil {
		log.Println("FATAL:Updating Dictation Complete [1]:" + err.Error())
	}

	sql = `
		UPDATE Dictations SET
			CompletedAt = datetime(current_timestamp, 'localtime')
		WHERE DocumentID = ?`

	_, err = app.sqliteWriter.Exec(sql, documentID)
	if err != nil {
		log.Println("FATAL:Updating Dictation Complete [2]:" + err.Error())
	}
}

func (app *App) DBAudioVaultUpdateSegmentMetadata(bitRate, duration, precision, sampleRate, filename string) {
	var sql = `
		UPDATE Segments SET
			AudioBitRate = ?,
			AudioDuration = ?,
			AudioPrecision = ?,
			AudioSampleRate = ?,
			ProcessingProgress = 1
		WHERE SegmentFileName = ?`

	_, err := app.sqliteWriter.Exec(sql, bitRate, duration, precision, sampleRate, filename)
	if err != nil {
		log.Println("FATAL:Updating Segments Audio Meta Data :" + err.Error())
	}
}

func (app *App) DBAudioVaultUpdateSegmentSoxReturnCode(filename string, code int) {
	var sql = `
		UPDATE Segments SET
			SoxStatusCode = ?
		WHERE SegmentFileName = ?`

	_, err := app.sqliteWriter.Exec(sql, code, filename)
	if err != nil {
		log.Println("FATAL:Updating Segments Sox Return Code :" + err.Error())
	}
}

func (app *App) DBAudioVaultInsertAuditEvent(filename, message string) {
	var sql = `
		INSERT INTO AuditEvents
			(EventAt, SegmentFileName, EventMessage)
			VALUES (datetime(current_timestamp, 'localtime'), ?, ?)`

	_, err := app.sqliteWriter.Exec(sql, filename, message)
	if err != nil {
		log.Println("FATAL:Inserting Audit Event : " + filename + " : " + message + " :" + err.Error())
	}
}

func (app *App) DBAudioVaultInsertDictation(submission *submission) {

	// fmt.Println("------------------------------")
	// fmt.Println(submission.DocumentID)
	// fmt.Println(submission.MRN)
	// fmt.Println(submission.CreatedBy)
	// fmt.Println(submission.MachineName)
	// fmt.Println(submission.SegmentCount)
	// fmt.Println(submission.SegmentOrder)
	// fmt.Println("------------------------------")

	var sql = `
		INSERT OR IGNORE INTO Dictations
			(DocumentID, MRN, CreatedBy, MachineName, SegmentCount, SavedAt)
			VALUES (?, ?, ?, ?, ?, datetime(current_timestamp, 'localtime'))`

	_, err := app.sqliteWriter.Exec(sql,
		submission.DocumentID,
		submission.MRN,
		submission.CreatedBy,
		submission.MachineName,
		submission.SegmentCount)

	if err != nil {
		log.Println("FATAL:Inserting Dictation : " + submission.DocumentID + " :" + err.Error())
	}
}

func (app *App) DBAudioVaultInsertSegment(submission *submission) {

	var sql = `
		INSERT OR IGNORE INTO Segments
			(SegmentFileName, DocumentID, SegmentFileSize, SegmentFileOrder, ProcessingProgress)
			VALUES (?, ?, ?, ?, 0)`

	_, err := app.sqliteWriter.Exec(sql,
		submission.SegmentFileName,
		submission.DocumentID,
		submission.SegmentFileSize,
		submission.SegmentOrder)

	if err != nil {
		log.Println("FATAL:Inserting Segment : " + submission.DocumentID + " :" + err.Error())
	}
}
