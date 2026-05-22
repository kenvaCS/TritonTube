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
