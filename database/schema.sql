-- tool to diagram the database schema 
-- https://dbdiagram.io/d

PRAGMA user_version = 1; -- https://www.sqlite.org/pragma.html#pragma_user_version
PRAGMA journal_mode = WAL; -- https://www.sqlite.org/pragma.html#pragma_journal_mode

DROP TABLE IF EXISTS Dictation;

CREATE TABLE Dictation (
    DocumentID          INTEGER PRIMARY KEY,
    MRN                 TEXT NOT NULL,
    DocumentName        TEXT NOT NULL DEFAULT '?'
);

CREATE INDEX idx_valid_mrn ON Dictation (MRN);

