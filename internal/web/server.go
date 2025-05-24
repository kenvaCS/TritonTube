// Lab 7: Implement a web server

package web

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type server struct {
	Addr string
	Port int

	metadataService VideoMetadataService
	contentService  VideoContentService

	mux *http.ServeMux
}

func NewServer(
	metadataService VideoMetadataService,
	contentService VideoContentService,
) *server {
	return &server{
		metadataService: metadataService,
		contentService:  contentService,
	}
}

func (s *server) Start(lis net.Listener) error {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/upload", s.handleUpload)
	s.mux.HandleFunc("/videos/", s.handleVideo)
	s.mux.HandleFunc("/content/", s.handleVideoContent)
	s.mux.HandleFunc("/", s.handleIndex)

	return http.Serve(lis, s.mux)
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// gather videos from metadata service
	videos, err := s.metadataService.List()
	if err != nil {
		http.Error(w, "Failed to gather videos", http.StatusInternalServerError)
		return
	}

	// have to get the escaped uris
	var data []struct {
		Id         string
		UploadTime string
		EscapedId  string
	}

	for _, video := range videos {
		data = append(data, struct {
			Id         string
			UploadTime string
			EscapedId  string
		}{
			Id:         video.Id,
			UploadTime: video.UploadedAt.Format("2006-01-02 15:04:05"),
			EscapedId:  url.PathEscape(video.Id),
		})
	}

	// create template if gathered videos and display it out to the page
	tmp := template.Must(template.New("index").Parse(indexHTML))
	if err := tmp.Execute(w, data); err != nil {
		log.Printf("Failed to render index template: %v", err)
		return
	}
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	// Open file submitted through HTTP
	f, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer f.Close()

	// Parse videoId from filename
	videoId := strings.TrimSuffix(header.Filename, ".mp4")
	if videoId == "" {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Upload Metadata of file sent (prior to lengthy processing of content)
	if err := s.metadataService.Create(videoId, time.Now()); err != nil {
		http.Error(w, "Failed to save video metadata", http.StatusInternalServerError)
		return
	}

	// Read in file data
	data, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, "Failed to read file content", http.StatusInternalServerError)
		return
	}

	// Create temp dir for mp4 file
	dir, err := os.MkdirTemp("", "storemp4")
	if err != nil {
		http.Error(w, "Failed to create temp directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(dir) // clean up

	// Create mp4 base file in temp dir
	videoPath := filepath.Join(dir, "video.mp4")
	v, err := os.Create(videoPath)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		return
	}

	_, err = io.Copy(v, bytes.NewReader(data))
	if err != nil {
		return
	}
	v.Close()

	// Create new temp dir for ffmpeg
	mpeg_dir, err := os.MkdirTemp("", "ffmpeg")
	if err != nil {
		http.Error(w, "Failed to create temp directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(mpeg_dir) // clean up

	// FFMpeg into this folder
	manifestPath := filepath.Join(mpeg_dir, "manifest.mpd")

	cmd := exec.Command("ffmpeg",
		"-i", videoPath, // input file
		"-c:v", "libx264", // video codec
		"-c:a", "aac", // audio codec
		"-bf", "1", // max 1 b-frame
		"-keyint_min", "120", // minimum keyframe interval
		"-g", "120", // keyframe every 120 frames
		"-sc_threshold", "0", // scene change threshold
		"-b:v", "3000k", // video bitrate
		"-b:a", "128k", // audio bitrate
		"-f", "dash", // dash format
		"-use_timeline", "1", // use timeline
		"-use_template", "1", // use template
		"-init_seg_name", "init-$RepresentationID$.m4s", // init segment naming
		"-media_seg_name", "chunk-$RepresentationID$-$Number%05d$.m4s", // media segment naming
		"-seg_duration", "4", // segment duration in seconds
		manifestPath) // output file

	if err := cmd.Run(); err != nil {
		log.Printf("ffmpeg failed: %v\n", err)
		return
	}

	// Iterate over ffmpeg temp dir, writing to content service
	files, err := os.ReadDir(mpeg_dir)
	if err != nil {
		http.Error(w, "Failed to read from temp dir", http.StatusInternalServerError)
	}

	for _, file := range files {
		if !file.IsDir() {
			fullPath := filepath.Join(mpeg_dir, file.Name())

			data, err := os.ReadFile(fullPath)
			if err != nil {
				log.Printf("Failed to read file %s: %v", fullPath, err)
				continue
			}

			if err := s.contentService.Write(videoId, file.Name(), data); err != nil {
				http.Error(w, "Failed to save part of video content", http.StatusInternalServerError)
				continue
			}
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *server) handleVideo(w http.ResponseWriter, r *http.Request) {
	videoId := r.URL.Path[len("/videos/"):]
	log.Println("Video ID:", videoId)

	video, _ := s.metadataService.Read(videoId)
	if video == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	tmp := template.Must(template.New("video").Parse(videoHTML))
	if err := tmp.Execute(w, video); err != nil {
		http.Error(w, "Failed to render video", http.StatusInternalServerError)
	}
}

func (s *server) handleVideoContent(w http.ResponseWriter, r *http.Request) {
	// parse /content/<videoId>/<filename>
	videoId := r.URL.Path[len("/content/"):]
	parts := strings.Split(videoId, "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid content path", http.StatusBadRequest)
		return
	}
	videoId = parts[0]
	filename := parts[1]
	log.Println("Video ID:", videoId, "Filename:", filename)

	data, err := s.contentService.Read(videoId, filename)
	if err != nil {
		http.Error(w, "Couldn't find content", http.StatusInternalServerError)
	}

	http.ServeContent(w, r, filename, time.Now(), bytes.NewReader(data))
}
