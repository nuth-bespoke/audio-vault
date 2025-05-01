-- tool to diagram the database schema 
-- https://dbdiagram.io/d

PRAGMA user_version = 1; -- https://www.sqlite.org/pragma.html#pragma_user_version
PRAGMA journal_mode = WAL; -- https://www.sqlite.org/pragma.html#pragma_journal_mode

DROP TABLE IF EXISTS AuditEvents;

CREATE TABLE AuditEvents (
    EventID             INTEGER PRIMARY KEY,
    EventAt             TEXT NOT NULL,
    SegmentFileName     TEXT NOT NULL,
    EventMessage        TEXT NOT NULL
);

CREATE INDEX idx_audit_events_filename ON AuditEvents (SegmentFileName, EventAt);


DROP TABLE IF EXISTS Dictations;

CREATE TABLE Dictations (
    DocumentID      INTEGER PRIMARY KEY,
    MRN             TEXT NOT NULL,
    DocumentName    TEXT NOT NULL DEFAULT '?',
    CreatedBy       TEXT NOT NULL,
    MachineName     TEXT NOT NULL,
    SavedAt         TEXT NOT NULL,
    SegmentCount    INTEGER NOT NULL DEFAULT 0,
    CompletedAt     TEXT,
    SentToDocstore  TEXT
);

CREATE INDEX idx_dictation_mrn ON Dictations (MRN);
CREATE INDEX idx_dictation_completed_at ON Dictations (CompletedAt);
CREATE INDEX idx_dictation_completed_at_docstore ON Dictations (CompletedAt, SentToDocstore);

DROP TABLE IF EXISTS Segments;

CREATE TABLE Segments (
    SegmentFileName     TEXT NOT NULL PRIMARY KEY,
    SavedAt             TEXT NOT NULL,
    DocumentID          INTEGER NOT NULL,
    SegmentFileSize     INTEGER NOT NULL,
    SegmentFileOrder    INTEGER NOT NULL,
    AudioBitRate        TEXT NOT NULL DEFAULT '?',
    AudioDuration       TEXT NOT NULL DEFAULT '?',
    AudioPrecision      TEXT NOT NULL DEFAULT '?',
    AudioSampleRate     TEXT NOT NULL DEFAULT '?',
    SoxStatusCode       INTEGER NOT NULL DEFAULT 0,
    ProcessingProgress  INTEGER NOT NULL DEFAULT 0, -- 0 = Not Processed
                                                    -- 1 = Meta Data Retrieved
                                                    -- 2 = Normalised Version Created
                                                    -- 3 = Segments Combined
    FOREIGN KEY(DocumentID) REFERENCES Dictations(DocumentID)
);

CREATE INDEX idx_segments_pending ON Segments 
    (ProcessingProgress, DocumentID, SegmentFileOrder)
    WHERE ProcessingProgress <= 2;

CREATE INDEX idx_segments_savedat ON Segments (SavedAt);



INSERT INTO Dictations (DocumentID, MRN, CreatedBy, MachineName, SavedAt, SegmentCount)
    VALUES (98767978, '0999994H', 'BRADLEYP6', 'P4X045', '2024-01-01 09:00:00', 2);

-- INSERT INTO Dictations (DocumentID, MRN, CreatedBy, MachineName, SavedAt, SegmentCount)
--     VALUES (98767970, '0999994H', 'BRADLEYP0', 'P4000', datetime(current_timestamp, 'localtime'), 2);

-- INSERT INTO Dictations (DocumentID, MRN, CreatedBy, MachineName, SavedAt, SegmentCount)
--     VALUES (999999900, '0999994H', 'BRADLEYP6', 'P4000', datetime(current_timestamp, 'localtime'), 3);
            
INSERT INTO Segments (SegmentFileName, DocumentID, SegmentFileSize, SegmentFileOrder, SavedAt)
    VALUES ('98767978-0999994H-12345-1.wav', 98767978, 567890, 1, '2024-01-01 09:00:00');

INSERT INTO Segments (SegmentFileName, DocumentID, SegmentFileSize, SegmentFileOrder, SavedAt)
    VALUES ('98767978-0999994H-67890-2.wav', 98767978, 55567890, 2, '2024-01-01 09:00:00');

-- INSERT INTO Segments (SegmentFileName, DocumentID, SegmentFileSize, SegmentFileOrder)
--     VALUES ('segment-1.wav', 999999900, 246606, 1);

-- INSERT INTO Segments (SegmentFileName, DocumentID, SegmentFileSize, SegmentFileOrder)
--     VALUES ('segment-2.wav', 999999900, 239694, 2);

-- INSERT INTO Segments (SegmentFileName, DocumentID, SegmentFileSize, SegmentFileOrder)
--     VALUES ('segment-3.wav', 999999900, 321486, 3);


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
       Segments.AudioSampleRate,
       Segments.ProcessingProgress
  FROM Segments
  LEFT JOIN Dictations ON Segments.DocumentID = Dictations.DocumentID
 WHERE ProcessingProgress <= 2
ORDER BY Segments.DocumentID, SegmentFileOrder;


--EXPLAIN QUERY PLAN
SELECT d.DocumentID, d.SegmentCount, COUNT(s.DocumentID) AS actual_segment_count
FROM Dictations d
LEFT JOIN Segments s ON d.DocumentID = s.DocumentID
WHERE d.CompletedAt IS NULL
GROUP BY d.DocumentID
HAVING COUNT(s.DocumentID) = d.SegmentCount AND ProcessingProgress = 0
LIMIT 0, 10;
