# Image Finder Web UI

Elegant web interface for the Face Recognition Image Finder.

## Features

- 🎨 Beautiful, modern UI
- 📤 Drag & drop file upload
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

The server will start on `http://localhost:8080`

### Usage

1. Open `http://localhost:8080` in your browser
2. Upload your ZIP archive
3. Upload 3-10 photos of yourself
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
