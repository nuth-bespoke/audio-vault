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
    DocumentID          INTEGER NOT NULL,
    SegmentFileSize     INTEGER NOT NULL,
    SegmentFileOrder    INTEGER NOT NULL,
    AudioBitRate        TEXT NOT NULL DEFAULT '?',
    AudioDuration       TEXT NOT NULL DEFAULT '?',
    AudioPrecision      TEXT NOT NULL DEFAULT '?',
    AudioSampleRate     TEXT NOT NULL DEFAULT '?',
    ProcessingProgress  INTEGER NOT NULL DEFAULT 0,

    FOREIGN KEY(DocumentID) REFERENCES Dictations(DocumentID)
);

CREATE INDEX idx_segments_pending ON Segments (ProcessingProgress, DocumentID, SegmentFileOrder);

INSERT INTO Dictations (DocumentID, MRN, CreatedBy, MachineName, SavedAt)
    VALUES (999, '7777h', 'MOSSXP', 'P4X045', DATE('now'));

INSERT INTO Segments (SegmentFileName, DocumentID, SegmentFileSize, SegmentFileOrder)
    VALUES ('999-7777h-12345-1.wav', 999, 567890, 1);

INSERT INTO Segments (SegmentFileName, DocumentID, SegmentFileSize, SegmentFileOrder)
    VALUES ('999-7777h-67890-2.wav', 999, 55567890, 2);

--EXPLAIN QUERY PLAN
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
ORDER BY Segments.DocumentID, SegmentFileOrder;


