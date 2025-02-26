package main

import (
	"bytes"
	"database/sql"
	"log"
	"os"
	"runtime"

	_ "modernc.org/sqlite"
)

func (app *App) DBAudioVaultClose() {
	app.sqliteReader.Close()
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
			Segments.AudioSampleRate
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
			&audioSegment.AudioSampleRate); err != nil {
			log.Println("ERR:" + err.Error())
		}

		segments.Segments = append(segments.Segments, audioSegment)
	}

	rows.Close()

	err = app.tplHTML.ExecuteTemplate(&tplBuffer, "segments-listing", segments)
	if err != nil {
		log.Println("ERR:" + err.Error())
	}

	return tplBuffer.String()
}

func (app *App) DBAudioVaultGetSegmentsPendingMetaData() string {
	var err error
	var rows *sql.Rows

	rows, err = app.sqliteReader.Query(`
		SELECT
			SegmentFileName
		 FROM Segments
		WHERE ProcessingProgress = 0
		ORDER BY DocumentID
		LIMIT 0, 10;`)
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
