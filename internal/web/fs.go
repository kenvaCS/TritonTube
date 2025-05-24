// Lab 7: Implement a local filesystem video content service

package web

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// FSVideoContentService implements VideoContentService using the local filesystem.
type FSVideoContentService struct {
	// probably want to initialize a directory where the filename should be stored/read from
	Directory string
}

// Uncomment the following line to ensure FSVideoContentService implements VideoContentService

func (s *FSVideoContentService) Read(videoId string, filename string) ([]byte, error) {
	file_path := filepath.Join(s.Directory, videoId, filename)
	f, err := os.Open(file_path)
	if err != nil {
		log.Printf("Error opening the file: %v", err)
		return nil, err
	}
	defer f.Close()

	// using loop here, but unsure if there's more idiomatic way to do it
	var data []byte
	buf := make([]byte, 1024)
	for {
		size, err := f.Read(buf)
		if size > 0 {
			data = append(data, buf[:size]...)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	return data, nil
}

func (s *FSVideoContentService) Write(videoId string, filename string, data []byte) error {
	// have to create directory
	err := os.MkdirAll(filepath.Join(s.Directory, videoId), 0750)
	if err != nil {
		log.Printf("Failed to create/find directory: %v\n", err)
		return err
	}

	// get file_path of destination in content_service
	file_path := filepath.Join(s.Directory, videoId, filename)
	f, err := os.Create(file_path)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		return err
	}
	defer f.Close()

	// need to explicitly close the file before executing ffmpeg-dash
	for len(data) > 0 {
		idx, err := f.Write(data)
		if err != nil {
			break
		}

		data = data[idx:]
	}

	return nil
}

var _ VideoContentService = (*FSVideoContentService)(nil) // assertion to ensure that implements VideoContentService
