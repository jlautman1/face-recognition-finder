# Image Finder Web UI

Elegant web interface for the Face Recognition Image Finder.

## Features

- 🎨 Beautiful, modern UI
- � Three input modes: Upload ZIP, Paste URL, Local Path
- 🔗 Server-side URL download (WeTransfer, Google Drive, Dropbox)
- 💻 Scan local folders or ZIPs already on the machine
- 📱 Mobile-friendly with camera capture for reference selfies
- 🌐 Accessible from any device on your local network
- 👤 Multiple reference photos support
- 🎚️ Adjustable matching threshold
- 📊 Real-time progress tracking
- ⬇️ Easy download of results

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Python 3.8–3.12 with the main script dependencies installed (Python 3.13+ deadlocks with TensorFlow/DeepFace)

### Installation

```bash
# Navigate to web directory
cd web

# Install Go dependencies
go mod download

# Run the server
go run main.go
```

The server will start on `http://0.0.0.0:8080` (accessible from other devices on your network)

### Usage

1. Open `http://localhost:8080` in your browser (or `http://<your-ip>:8080` from a phone)
2. Choose image source: Upload ZIP, Paste URL, or Local Path
3. Upload 3-10 photos of yourself (on mobile, the front camera opens directly)
4. Adjust the matching threshold (optional)
5. Click "Start Processing"
6. Wait for results and download!

## Architecture

- **Backend**: Go HTTP server
- **Frontend**: Vanilla HTML/CSS/JavaScript
- **Processing**: Calls Python script with subprocess
- **Storage**: Temporary files in `uploads/` and `outputs/`

## API Endpoints

- `GET /` - Serve main page
- `POST /upload` - Upload files and start processing
- `GET /status/:jobId` - Check job status
- `GET /download/:jobId` - Download results

## Configuration

Edit `main.go` to change:
- Port number (default: 8080)
- Upload size limits
- Directories
- Python script path
