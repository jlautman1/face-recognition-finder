package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	uploadsDir   = "./uploads"
	outputsDir   = "./outputs"
	pythonScript = "../find_my_images.py"
)

type JobStatus struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // pending, processing, completed, error
	Progress  int       `json:"progress"`
	Message   string    `json:"message"`
	Total     int       `json:"total"`
	Scanned   int       `json:"scanned"`
	Matches   int       `json:"matches"`
	StartTime time.Time `json:"start_time"`
	mu        sync.RWMutex
}

var jobs = make(map[string]*JobStatus)

// Thread-safe methods for JobStatus
func (j *JobStatus) Update(fn func()) {
	j.mu.Lock()
	defer j.mu.Unlock()
	fn()
}

func (j *JobStatus) GetSnapshot() JobStatus {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return JobStatus{
		ID:        j.ID,
		Status:    j.Status,
		Progress:  j.Progress,
		Message:   j.Message,
		Total:     j.Total,
		Scanned:   j.Scanned,
		Matches:   j.Matches,
		StartTime: j.StartTime,
	}
}

func main() {
	// Create necessary directories
	os.MkdirAll(uploadsDir, 0755)
	os.MkdirAll(outputsDir, 0755)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API endpoints
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/status/", handleStatus)
	http.HandleFunc("/download/", handleDownload)

	port := ":8080"
	log.Printf("🚀 Server starting on http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10GB)
	err := r.ParseMultipartForm(10 << 30)
	if err != nil {
		respondError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get threshold
	threshold := r.FormValue("threshold")
	if threshold == "" {
		threshold = "0.35"
	}

	// Create job ID
	jobID := uuid.New().String()
	jobDir := filepath.Join(uploadsDir, jobID)
	refDir := filepath.Join(jobDir, "references")
	outputDir := filepath.Join(outputsDir, jobID)

	os.MkdirAll(refDir, 0755)
	os.MkdirAll(outputDir, 0755)

	// Initialize job status
	jobs[jobID] = &JobStatus{
		ID:        jobID,
		Status:    "pending",
		Progress:  0,
		Message:   "Uploading files...",
		StartTime: time.Now(),
	}

	// Save ZIP file
	zipFile, zipHeader, err := r.FormFile("zipFile")
	if err != nil {
		respondError(w, "ZIP file is required", http.StatusBadRequest)
		return
	}
	defer zipFile.Close()

	zipPath := filepath.Join(jobDir, zipHeader.Filename)
	dst, err := os.Create(zipPath)
	if err != nil {
		respondError(w, "Failed to save ZIP file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, zipFile)

	// Save reference images
	refFiles := r.MultipartForm.File["refImages"]
	if len(refFiles) == 0 {
		respondError(w, "At least one reference image is required", http.StatusBadRequest)
		return
	}

	for _, fileHeader := range refFiles {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		refPath := filepath.Join(refDir, fileHeader.Filename)
		dst, err := os.Create(refPath)
		if err != nil {
			continue
		}
		defer dst.Close()
		io.Copy(dst, file)
	}

	// Start processing in background
	go processImages(jobID, zipPath, refDir, outputDir, threshold)

	// Return job ID
	json.NewEncoder(w).Encode(map[string]string{
		"job_id":  jobID,
		"message": "Processing started",
	})
}

func processImages(jobID, zipPath, refDir, outputDir, threshold string) {
	job := jobs[jobID]
	job.Update(func() {
		job.Status = "processing"
		job.Message = "Processing images..."
	})

	// Build Python command
	pythonCmd := "python3"
	// Check if venv exists
	venvPython := "../venv/bin/python"
	if _, err := os.Stat(venvPython); err == nil {
		pythonCmd = venvPython
	}

	// Run Python with -u flag for unbuffered output to get real-time updates
	cmd := exec.Command(pythonCmd, "-u", pythonScript, zipPath, refDir, outputDir, threshold)

	// Create pipes to stream output in real-time
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		job.Update(func() {
			job.Status = "error"
			job.Message = fmt.Sprintf("Error creating stdout pipe: %v", err)
		})
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		job.Update(func() {
			job.Status = "error"
			job.Message = fmt.Sprintf("Error creating stderr pipe: %v", err)
		})
		return
	}

	// Start command
	if err := cmd.Start(); err != nil {
		job.Update(func() {
			job.Status = "error"
			job.Message = fmt.Sprintf("Error starting process: %v", err)
		})
		return
	}

	// Read output in real-time
	scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line) // Log for debugging

		// Parse progress: [5/100] Scanned...
		if strings.Contains(line, "[") && strings.Contains(line, "/") && strings.Contains(line, "]") {
			start := strings.Index(line, "[")
			end := strings.Index(line, "]")
			if start >= 0 && end > start {
				progress := line[start+1 : end]
				parts := strings.Split(progress, "/")
				if len(parts) == 2 {
					var current, total int
					fmt.Sscanf(parts[0], "%d", &current)
					fmt.Sscanf(parts[1], "%d", &total)
					if total > 0 {
						job.Update(func() {
							job.Progress = (current * 100) / total
							job.Total = total
							job.Scanned = current
						})
					}
				}
			}
		}

		// Parse matches: "MATCH" or "matches so far"
		if strings.Contains(line, "✓ MATCH") {
			job.Update(func() {
				job.Matches++
			})
		} else if strings.Contains(line, "matches so far") {
			// Extract number from "(X matches so far)"
			start := strings.Index(line, "(")
			end := strings.Index(line, " matches")
			if start >= 0 && end > start {
				matchStr := line[start+1 : end]
				var matches int
				fmt.Sscanf(matchStr, "%d", &matches)
				job.Update(func() {
					job.Matches = matches
				})
			}
		}

		// Parse final stats
		if strings.Contains(line, "Matches found:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				matchesStr := strings.TrimSpace(parts[1])
				var matches int
				fmt.Sscanf(matchesStr, "%d", &matches)
				job.Update(func() {
					job.Matches = matches
				})
			}
		}

		// Update message
		if strings.Contains(line, "Scanned") || strings.Contains(line, "MATCH") {
			job.Update(func() {
				job.Message = line
			})
		}
	}

	// Wait for command to complete
	err = cmd.Wait()

	if err != nil {
		job.Update(func() {
			job.Status = "error"
			job.Message = fmt.Sprintf("Process error: %v", err)
		})
		return
	}

	job.Update(func() {
		job.Status = "completed"
		job.Progress = 100
		job.Message = fmt.Sprintf("Processing complete! Found %d matching images.", job.Matches)
	})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	jobID := strings.TrimPrefix(r.URL.Path, "/status/")

	job, exists := jobs[jobID]
	if !exists {
		respondError(w, "Job not found", http.StatusNotFound)
		return
	}

	// Get a thread-safe snapshot of the job status
	snapshot := job.GetSnapshot()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	jobID := strings.TrimPrefix(r.URL.Path, "/download/")

	job, exists := jobs[jobID]
	if !exists {
		respondError(w, "Job not found", http.StatusNotFound)
		return
	}

	if job.Status != "completed" {
		respondError(w, "Job not completed yet", http.StatusBadRequest)
		return
	}

	outputDir := filepath.Join(outputsDir, jobID)

	// Check if output directory exists and has files
	files, err := os.ReadDir(outputDir)
	if err != nil || len(files) == 0 {
		respondError(w, "No results found", http.StatusNotFound)
		return
	}

	// Create a ZIP of the output folder
	zipPath := filepath.Join(outputsDir, jobID+".zip")

	// Use absolute paths for zip command
	absZipPath, _ := filepath.Abs(zipPath)
	absOutputDir, _ := filepath.Abs(outputDir)

	// Create zip with all files in the output directory
	cmd := exec.Command("zip", "-j", "-r", absZipPath, absOutputDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("ZIP creation error: %v, output: %s", err, string(output))
		respondError(w, "Failed to create download archive", http.StatusInternalServerError)
		return
	}

	// Serve the ZIP file
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=matches-%s.zip", jobID))
	http.ServeFile(w, r, absZipPath)
}

func respondError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
