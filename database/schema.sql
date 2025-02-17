-- tool to diagram the database schema 
-- https://dbdiagram.io/d

PRAGMA user_version = 1; -- https://www.sqlite.org/pragma.html#pragma_user_version
PRAGMA journal_mode = WAL; -- https://www.sqlite.org/pragma.html#pragma_journal_mode

DROP TABLE IF EXISTS Dictations;

CREATE TABLE Dictations (
    DocumentID      INTEGER PRIMARY KEY,
    MRN             TEXT NOT NULL,
    DocumentName    TEXT NOT NULL DEFAULT '?',
    CreatedBy       TEXT NOT NULL,
    MachineName     TEXT NOT NULL,
    SavedAt         TEXT NOT NULL
);

CREATE INDEX idx_dictation_mrn ON Dictations (MRN);


DROP TABLE IF EXISTS Segments;

CREATE TABLE Segments (
    SegmentFileName     TEXT NOT NULL PRIMARY KEY,
    SegmentFileSize     INTEGER NOT NULL,
    SegmentFileOrder    INTEGER NOT NULL,
    AudioBitRate        TEXT NOT NULL,
    AudioDuration       TEXT NOT NULL,
    AudioPrecision      TEXT NOT NULL,
    AudioSampleRate     TEXT NOT NULL
);

