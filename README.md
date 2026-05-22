[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/e5W8wwsN)

# TritonTube

A distributed video streaming platform inspired by YouTube, built in Go. Users can upload MP4 videos which are transcoded into **MPEG-DASH** adaptive bitrate format and served for in-browser playback.

## Architecture

The system is composed of three services:

| Service | Description |
|---|---|
| `cmd/web` | HTTP frontend — handles video upload, listing, and playback |
| `cmd/storage` | Distributed storage node for video content |
| `cmd/admin` | gRPC admin service for managing storage cluster nodes |

## Tech Stack

- **Go 1.24** — core language
- **FFmpeg** — transcodes uploaded `.mp4` files to MPEG-DASH (`.mpd` + `.m4s` segments)
- **SQLite** (`go-sqlite3`) — video metadata persistence
- **gRPC + Protocol Buffers** — inter-service communication for storage node management
- **Consistent hashing** — distributes video content across storage nodes
- **MPEG-DASH** — adaptive bitrate video streaming in the browser
