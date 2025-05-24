// Lab 7: Implement a SQLite video metadata service

package web

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // blank identifier import, what does it do tho?
)

/* Assumptions/Basis made:
 * CREATE TABLE videos (id TEXT PRIMARY KEY, uploaded_at DATETIME NOT NULL); was the observed SQL table schema*
 *
 */

type SQLiteVideoMetadataService struct {
	// what else does this need?
	DB *sql.DB // important that DB is capitalized (at least D) otherwise is private and needs some constructor to initialize/methods
}

// specifies "receiver" so have access to internal variables of "struct" we're calling from
func (s *SQLiteVideoMetadataService) Create(videoId string, uploadedAt time.Time) error { // save new entry

	// have to convert uploadedAt into a format for lossless conversion to DATETIME in sql
	dbUploadedAt := uploadedAt.Format(time.RFC3339)
	_, err := s.DB.Exec("INSERT INTO videos (id, uploaded_at) VALUES (?, ?)", videoId, dbUploadedAt)
	if err != nil {
		log.Printf("failed attempt to create new metadata record %v\n", err)
		return err
	}

	return nil
}

func (s *SQLiteVideoMetadataService) List() ([]VideoMetadata, error) {
	rows, err := s.DB.Query("SELECT id, uploaded_at FROM videos")
	if err != nil {
		log.Printf("Failure obtaining rows from query %v\n", err)
		return nil, err // can't read anything if error reading rows
	}
	defer rows.Close() // only relevant if rows exists to be closed, meaning no error occured

	var ret []VideoMetadata
	for rows.Next() {
		var videoId string
		var stringUploaded string
		var uploadedAt time.Time

		if err := rows.Scan(&videoId, &stringUploaded); err != nil {
			// not sure best way to handle it if there's an error, maybe just to return the current list and log an error?
			log.Printf("Error occured when reading out all records %v\n", err)
			return ret, nil
		}
		// have to convert back to an actual time.Time() typed variable
		uploadedAt, err := time.Parse(time.RFC3339, stringUploaded)
		if err != nil {
			log.Printf("Failed to parse uploaded_at time: %v\n", err)
			continue
		}
		ret = append(ret, VideoMetadata{Id: videoId, UploadedAt: uploadedAt}) // don't need to unpack because
	}

	return ret, nil
}

func (s *SQLiteVideoMetadataService) Read(videoId string) (*VideoMetadata, error) { // specific entry lookup
	/* QueryRow since guaranteed each video has unique Id. Also row contains
	possible error value, so can just poll it's validity when we call Scan */
	row := s.DB.QueryRow("SELECT id, uploaded_at FROM videos WHERE id = ?", videoId)

	// Probably best practice to use the value obtained from db for consistency/integrity
	var dbVideoId string
	var stringUploaded string

	if err := row.Scan(&dbVideoId, &stringUploaded); err != nil {
		log.Printf("Failure to find and read out specific entry %v\n", err)
		return nil, err
	}

	// same thing here as in list
	uploadedAt, err := time.Parse(time.RFC3339, stringUploaded)
	if err != nil {
		log.Printf("Failed to parse uploaded_at time: %v\n", err)
		return nil, err
	}

	return &VideoMetadata{Id: dbVideoId, UploadedAt: uploadedAt}, nil
}

var _ VideoMetadataService = (*SQLiteVideoMetadataService)(nil) // compile time check
