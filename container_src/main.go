package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	message := os.Getenv("MESSAGE")
	instanceId := os.Getenv("CLOUDFLARE_DURABLE_OBJECT_ID")
	fmt.Fprintf(w, "Hi, I'm a container and this is my message: \"%s\", my instance ID is: %s", message, instanceId)

}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	panic("This is a panic")
}

type FFmpegRequest struct {
	VideoURL string `json:"video_url"`
}

type FFmpegResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	AudioURL  string `json:"audio_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

func ffmpegHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req FFmpegRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := FFmpegResponse{
			Success: false,
			Error:   "Invalid JSON payload",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.VideoURL == "" {
		response := FFmpegResponse{
			Success: false,
			Error:   "video_url is required",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create unique filenames
	instanceId := os.Getenv("CLOUDFLARE_DURABLE_OBJECT_ID")
	if instanceId == "" {
		instanceId = "default"
	}
	
	videoFile := filepath.Join("/tmp/processing", fmt.Sprintf("video_%s_%d.tmp", instanceId, time.Now().Unix()))
	audioFile := filepath.Join("/tmp/processing", fmt.Sprintf("audio_%s_%d.mp3", instanceId, time.Now().Unix()))

	// Download the video file
	resp, err := http.Get(req.VideoURL)
	if err != nil {
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to download video: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to download video: HTTP %d", resp.StatusCode),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Save video to temporary file
	videoFileHandle, err := os.Create(videoFile)
	if err != nil {
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temp file: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer videoFileHandle.Close()
	defer os.Remove(videoFile) // Clean up

	_, err = io.Copy(videoFileHandle, resp.Body)
	if err != nil {
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to save video: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Extract audio using FFmpeg
	cmd := exec.Command("ffmpeg", "-i", videoFile, "-vn", "-acodec", "mp3", "-ab", "192k", "-ar", "44100", "-y", audioFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("FFmpeg failed: %v, output: %s", err, string(output)),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if audio file was created
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		response := FFmpegResponse{
			Success: false,
			Error:   "Audio file was not created",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// For this example, we'll return success with the local file path
	// In a production environment, you'd want to upload to cloud storage
	response := FFmpegResponse{
		Success:  true,
		Message:  "Audio extracted successfully",
		AudioURL: fmt.Sprintf("/download/%s", filepath.Base(audioFile)),
	}

	json.NewEncoder(w).Encode(response)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/download/"):]
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join("/tmp/processing", filename)
	
	// Security check - ensure file is in our processing directory
	if !filepath.HasPrefix(filePath, "/tmp/processing/") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set headers for MP3 download
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Serve the file
	http.ServeFile(w, r, filePath)

	// Clean up file after serving (optional)
	go func() {
		time.Sleep(5 * time.Minute) // Give some time for download
		os.Remove(filePath)
	}()
}

func main() {
	// Listen for SIGINT and SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	router := http.NewServeMux()
	router.HandleFunc("/", handler)
	router.HandleFunc("/container", handler)
	router.HandleFunc("/error", errorHandler)
	router.HandleFunc("/ffmpeg/extract-audio", ffmpegHandler)
	router.HandleFunc("/download/", downloadHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Wait to receive a signal
	sig := <-stop

	log.Printf("Received signal (%s), shutting down server...", sig)

	// Give the server 5 seconds to shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server shutdown successfully")
}
