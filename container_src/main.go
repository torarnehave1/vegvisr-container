package main

import (
	"context"
	"encoding/base64"
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

func readmeHandler(w http.ResponseWriter, r *http.Request) {
	// GitHub raw content URL for README.md
	githubURL := "https://raw.githubusercontent.com/torarnehave1/vegvisr-container/main/README.md"
	
	log.Printf("Fetching README.md from GitHub: %s", githubURL)
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Fetch README from GitHub
	resp, err := client.Get(githubURL)
	if err != nil {
		log.Printf("Failed to fetch README from GitHub: %v", err)
		http.Error(w, fmt.Sprintf("Failed to fetch README: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("GitHub returned non-200 status: %d", resp.StatusCode)
		http.Error(w, fmt.Sprintf("GitHub returned status: %d", resp.StatusCode), resp.StatusCode)
		return
	}
	
	// Set headers for markdown file download
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"README.md\"")
	
	// Copy the content directly from GitHub to the response
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Failed to copy README content: %v", err)
		return
	}
	
	log.Printf("Successfully served README.md from GitHub")
}

// uploadBase64Handler handles file uploads sent as base64 JSON from the TypeScript worker
func uploadBase64Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Upload base64 handler called")
	
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Parse JSON request
	var req struct {
		VideoData    string `json:"video_data"`
		Filename     string `json:"filename"`
		FileSize     int64  `json:"file_size"`
		OutputFormat string `json:"output_format"`
		InstanceId   string `json:"instance_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Failed to parse JSON request: %v", err)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Received base64 upload: %s, size: %d bytes (%.2f MB)", req.Filename, req.FileSize, float64(req.FileSize)/(1024*1024))

	// Check file size (50MB limit for testing)
	const maxFileSize = 50 * 1024 * 1024 // 50MB
	if req.FileSize > maxFileSize {
		log.Printf("File too large: %d bytes (max: %d)", req.FileSize, maxFileSize)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("File too large (%.1f MB). Maximum supported: 50MB", float64(req.FileSize)/(1024*1024)),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate output format
	outputFormat := req.OutputFormat
	if outputFormat == "" {
		outputFormat = "mp3"
	}
	validFormats := map[string]bool{"mp3": true, "wav": true, "aac": true, "flac": true}
	if !validFormats[outputFormat] {
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Unsupported output format: %s. Supported: mp3, wav, aac, flac", outputFormat),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	instanceId := req.InstanceId
	if instanceId == "" {
		instanceId = "upload"
	}

	// Decode base64 video data
	log.Printf("Decoding base64 video data...")
	videoData, err := base64.StdEncoding.DecodeString(req.VideoData)
	if err != nil {
		log.Printf("Failed to decode base64 video data: %v", err)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to decode video data: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create temp directory
	tempDir := "/tmp/processing"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("Failed to create temp directory: %v", err)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temp directory: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate filenames
	timestamp := time.Now().UnixMilli()
	videoFile := filepath.Join(tempDir, fmt.Sprintf("upload_%s_%d%s", instanceId, timestamp, filepath.Ext(req.Filename)))
	audioFile := filepath.Join(tempDir, fmt.Sprintf("audio_%s_%d.%s", instanceId, timestamp, outputFormat))

	log.Printf("Saving video to: %s", videoFile)

	// Save video data to file
	err = os.WriteFile(videoFile, videoData, 0644)
	if err != nil {
		log.Printf("Failed to save video file: %v", err)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to save video file: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer os.Remove(videoFile) // Clean up

	log.Printf("Video saved successfully, starting FFmpeg processing")

	// Process with FFmpeg
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Starting FFmpeg processing with %s format (file size: %.2f MB), timeout: 30s", 
		outputFormat, float64(req.FileSize)/(1024*1024))

	var cmd *exec.Cmd
	switch outputFormat {
	case "wav":
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "pcm_s16le", "-ar", "44100", "-y", audioFile)
	case "flac":
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "flac", "-ar", "44100", "-y", audioFile)
	case "aac":
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "aac", "-ab", "192k", "-ar", "44100", "-y", audioFile)
	default: // mp3
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "mp3", "-ab", "192k", "-ar", "44100", "-y", audioFile)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("FFmpeg failed: %v, output: %s", err, string(output))
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("ffmpeg failed: %v, stderr: %s", err, string(output)),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("FFmpeg processing completed successfully")

	// Read processed audio and return as base64
	audioInfo, err := os.Stat(audioFile)
	if err != nil {
		log.Printf("Failed to get audio file info: %v", err)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get audio file info: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	audioSize := audioInfo.Size()
	audioSizeMB := float64(audioSize) / (1024 * 1024)
	audioFileName := fmt.Sprintf("audio_%s_%d.%s", instanceId, timestamp, outputFormat)
	
	log.Printf("Processed audio file: %s (size: %.2f MB)", audioFileName, audioSizeMB)
	
	// Read and encode audio file
	audioData, err := os.ReadFile(audioFile)
	if err != nil {
		log.Printf("Failed to read processed audio file: %v", err)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read processed audio: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer os.Remove(audioFile) // Clean up
	
	// Encode audio data as base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)
	
	response := FFmpegResponse{
		Success:     true,
		Message:     "Audio extracted from uploaded file successfully",
		DownloadURL: fmt.Sprintf("/download/%s", audioFileName),
		FileName:    audioFileName,
		R2Key:       audioFileName,
		AudioURL:    fmt.Sprintf("/download/%s", audioFileName),
		AudioData:   audioBase64,
	}
	
	log.Printf("Upload processing completed: %s", audioFileName)
	json.NewEncoder(w).Encode(response)
}

// uploadHandler handles direct file uploads from frontend
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Upload handler called")
	
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form (64MB max memory to handle larger files)
	log.Printf("Starting multipart form parsing...")
	err := r.ParseMultipartForm(64 << 20) // 64MB
	if err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}
	log.Printf("Multipart form parsed successfully")

	// Get the uploaded file
	log.Printf("Getting file from form...")
	file, header, err := r.FormFile("video")
	if err != nil {
		log.Printf("Failed to get file from form: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get uploaded file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file upload: %s, size: %d bytes (%.2f MB)", header.Filename, header.Size, float64(header.Size)/(1024*1024))

	// Check file size (200MB limit)
	const maxFileSize = 200 * 1024 * 1024 // 200MB
	if header.Size > maxFileSize {
		log.Printf("File too large: %d bytes (max: %d)", header.Size, maxFileSize)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("File too large (%.1f MB). Maximum supported: 200MB", float64(header.Size)/(1024*1024)),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get output format from form (default to mp3)
	outputFormat := r.FormValue("output_format")
	if outputFormat == "" {
		outputFormat = "mp3"
	}

	// Validate output format
	validFormats := map[string]bool{"mp3": true, "wav": true, "aac": true, "flac": true}
	if !validFormats[outputFormat] {
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Unsupported output format: %s. Supported: mp3, wav, aac, flac", outputFormat),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get instance ID from URL path or form
	instanceId := r.FormValue("instance_id")
	if instanceId == "" {
		instanceId = "upload"
	}

	// Create temp directory
	tempDir := "/tmp/processing"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("Failed to create temp directory: %v", err)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temp directory: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate filenames
	timestamp := time.Now().UnixMilli()
	videoFile := filepath.Join(tempDir, fmt.Sprintf("upload_%s_%d%s", instanceId, timestamp, filepath.Ext(header.Filename)))
	audioFile := filepath.Join(tempDir, fmt.Sprintf("audio_%s_%d.%s", instanceId, timestamp, outputFormat))

	log.Printf("Saving uploaded file to: %s", videoFile)

	// Save uploaded file to temp location
	videoFileHandle, err := os.Create(videoFile)
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temp file: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	defer videoFileHandle.Close()
	defer os.Remove(videoFile) // Clean up

	// Copy uploaded file to temp location with progress logging
	log.Printf("Starting file copy operation...")
	bytesWritten, err := io.Copy(videoFileHandle, file)
	if err != nil {
		log.Printf("Failed to save uploaded file: %v", err)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to save uploaded file: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	videoFileHandle.Close() // Close before FFmpeg processing

	log.Printf("File saved successfully! Wrote %d bytes, starting FFmpeg processing", bytesWritten)

	// Process with FFmpeg using same logic as URL-based processing
	w.Header().Set("Content-Type", "application/json")

	// Process the uploaded video file with FFmpeg
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Shorter timeout for testing
	defer cancel()

	log.Printf("Starting FFmpeg processing with %s format (file size: %.2f MB), timeout: 30s", 
		outputFormat, float64(header.Size)/(1024*1024))

	var cmd *exec.Cmd
	switch outputFormat {
	case "wav":
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "pcm_s16le", "-ar", "44100", "-y", audioFile)
	case "flac":
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "flac", "-ar", "44100", "-y", audioFile)
	case "aac":
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "aac", "-ab", "192k", "-ar", "44100", "-y", audioFile)
	default: // mp3
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "mp3", "-ab", "192k", "-ar", "44100", "-y", audioFile)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("FFmpeg failed: %v, output: %s", err, string(output))
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("ffmpeg failed: %v, stderr: %s", err, string(output)),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("FFmpeg processing completed successfully")

	// Read the processed audio file and encode as base64 for R2 upload
	// But only for smaller files to avoid Cloudflare limits
	audioInfo, err := os.Stat(audioFile)
	if err != nil {
		log.Printf("Failed to get audio file info: %v", err)
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get audio file info: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	audioSize := audioInfo.Size()
	audioSizeMB := float64(audioSize) / (1024 * 1024)
	audioFileName := fmt.Sprintf("audio_%s_%d.%s", instanceId, timestamp, outputFormat)
	
	log.Printf("Processed audio file: %s (size: %.2f MB)", audioFileName, audioSizeMB)
	
	// If audio file is small enough (< 10MB), include it in response for R2 upload
	// Otherwise, we'll need to implement direct R2 upload from container
	if audioSize < 10*1024*1024 { // 10MB limit
		audioData, err := os.ReadFile(audioFile)
		if err != nil {
			log.Printf("Failed to read processed audio file: %v", err)
			response := FFmpegResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to read processed audio: %v", err),
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		
		// Encode audio data as base64 for transfer
		audioBase64 := base64.StdEncoding.EncodeToString(audioData)
		
		response := FFmpegResponse{
			Success:     true,
			Message:     "Audio extracted from uploaded file and ready for storage",
			DownloadURL: fmt.Sprintf("/download/%s", audioFileName),
			FileName:    audioFileName,
			R2Key:       audioFileName,
			AudioURL:    fmt.Sprintf("/download/%s", audioFileName),
			AudioData:   audioBase64,
		}
		
		log.Printf("Upload processing completed: %s (included in response)", audioFileName)
		json.NewEncoder(w).Encode(response)
	} else {
		// File too large for response - need direct R2 upload or return error
		response := FFmpegResponse{
			Success: false,
			Error:   fmt.Sprintf("Processed audio too large (%.1f MB) for direct upload. Use URL-based processing for large files.", audioSizeMB),
		}
		json.NewEncoder(w).Encode(response)
	}
	
	// Clean up the audio file
	defer os.Remove(audioFile)
}

// ProgressCallback is a function type for progress updates
type ProgressCallback func(stage, message string, progress float64)

// downloadDirectURL downloads a video from a direct URL with chunked downloading and progress updates
func downloadDirectURLWithProgress(url, outputPath string, progressCallback ProgressCallback) error {
	log.Printf("Starting chunked download from: %s", url)
	log.Printf("Target path: %s", outputPath)
	
	progressCallback("info", "Getting file information...", 0)
	
	// First, get the file size with a HEAD request
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	headResp, err := client.Head(url)
	if err != nil {
		log.Printf("HEAD request failed: %v", err)
		return fmt.Errorf("failed to get file info: %v", err)
	}
	
	fileSize := headResp.ContentLength
	fileSizeMB := float64(fileSize) / (1024 * 1024)
	log.Printf("File size: %d bytes (%.2f MB)", fileSize, fileSizeMB)
	
	progressCallback("info", fmt.Sprintf("File size: %.2f MB", fileSizeMB), 5)
	
	// Check if file is too large (chunking allows much larger files)
	if fileSize > 200*1024*1024 { // 200MB limit - chunked download makes this feasible
		return fmt.Errorf("video file too large (%.1f MB). Maximum supported: 200MB", fileSizeMB)
	}
	
	// Create output file
	videoFileHandle, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer videoFileHandle.Close()
	
	// Download in 5MB chunks
	chunkSize := int64(5 * 1024 * 1024) // 5MB chunks
	var totalWritten int64
	totalChunks := (fileSize + chunkSize - 1) / chunkSize
	
	progressCallback("download", "Starting chunked download...", 10)
	
	for chunkNum, start := int64(1), int64(0); start < fileSize; chunkNum, start = chunkNum+1, start+chunkSize {
		end := start + chunkSize - 1
		if end >= fileSize {
			end = fileSize - 1
		}
		
		chunkSizeMB := float64(end-start+1) / (1024 * 1024)
		progressMsg := fmt.Sprintf("Downloading chunk %d/%d (%.1f MB)", chunkNum, totalChunks, chunkSizeMB)
		progress := 10 + (float64(chunkNum-1)/float64(totalChunks))*50 // 10-60% for download
		
		progressCallback("download", progressMsg, progress)
		log.Printf("Downloading chunk %d/%d: %d-%d", chunkNum, totalChunks, start, end)
		
		// Create range request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create range request: %v", err)
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
		
		// Download chunk with timeout
		chunkClient := &http.Client{
			Timeout: 15 * time.Second, // Longer timeout for larger chunks
		}
		
		resp, err := chunkClient.Do(req)
		if err != nil {
			log.Printf("Chunk download failed: %v", err)
			return fmt.Errorf("failed to download chunk %d: %v", chunkNum, err)
		}
		
		if resp.StatusCode != 206 && resp.StatusCode != 200 { // 206 = Partial Content
			resp.Body.Close()
			return fmt.Errorf("server doesn't support range requests or failed: HTTP %d", resp.StatusCode)
		}
		
		// Copy chunk to file
		written, err := io.Copy(videoFileHandle, resp.Body)
		resp.Body.Close()
		
		if err != nil {
			log.Printf("Failed to write chunk %d: %v", chunkNum, err)
			return fmt.Errorf("failed to write chunk %d: %v", chunkNum, err)
		}
		
		totalWritten += written
		completedPct := float64(totalWritten) / float64(fileSize) * 100
		log.Printf("Chunk %d written: %d bytes (total: %.1f%% - %d/%d bytes)", 
			chunkNum, written, completedPct, totalWritten, fileSize)
	}
	
	progressCallback("download", "Download completed!", 60)
	log.Printf("Download completed: %d bytes written", totalWritten)
	return nil
}

// Backward compatibility wrapper
func downloadDirectURL(url, outputPath string) error {
	return downloadDirectURLWithProgress(url, outputPath, func(stage, message string, progress float64) {
		// Silent progress - just log
		log.Printf("[%s] %.1f%% - %s", stage, progress, message)
	})
}

type FFmpegRequest struct {
	VideoURL     string `json:"video_url"`
	UseR2Storage bool   `json:"use_r2_storage"`
	InstanceID   string `json:"instance_id"`
	AudioFormat  string `json:"audio_format,omitempty"` // mp3, wav, etc.
	AudioQuality string `json:"audio_quality,omitempty"` // 192k, 320k, etc.
}

type FFmpegResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	AudioData     string `json:"audio_data,omitempty"`
	AudioURL      string `json:"audio_url,omitempty"`
	DownloadURL   string `json:"download_url,omitempty"`
	FileName      string `json:"file_name,omitempty"`
	R2Key         string `json:"r2_key,omitempty"`
	Error         string `json:"error,omitempty"`
	VideoTitle    string `json:"video_title,omitempty"`
	Duration      string `json:"duration,omitempty"`
	VideoSource   string `json:"video_source,omitempty"`
	Progress      string `json:"progress,omitempty"`
	FileSize      string `json:"file_size,omitempty"`
	DownloadSpeed string `json:"download_speed,omitempty"`
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
	
	timestamp := time.Now().Unix()
	videoFile := filepath.Join("/tmp/processing", fmt.Sprintf("video_%s_%d.tmp", instanceId, timestamp))
	
	// Set audio format and quality defaults
	audioFormat := req.AudioFormat
	if audioFormat == "" {
		audioFormat = "mp3"
	}
	
	audioQuality := req.AudioQuality
	if audioQuality == "" {
		audioQuality = "192k"
	}
	
	audioFile := filepath.Join("/tmp/processing", fmt.Sprintf("audio_%s_%d.%s", instanceId, timestamp, audioFormat))

	var videoTitle, duration, videoSource string

	// Download from direct URL with progress updates
	log.Printf("Downloading from URL: %s", req.VideoURL)
	videoSource = "direct"
	
	// Send initial progress response
	w.Header().Set("Content-Type", "application/json")
	
	progressCallback := func(stage, message string, progress float64) {
		// For now, just log progress. In a real implementation, you might use Server-Sent Events
		log.Printf("Progress: %.1f%% - %s", progress, message)
	}
	
	err := downloadDirectURLWithProgress(req.VideoURL, videoFile, progressCallback)
	if err != nil {
		response := FFmpegResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Clean up video file after processing
	defer os.Remove(videoFile)

	// Extract audio using FFmpeg with configurable format and quality
	// Use context with timeout for FFmpeg processing (longer for larger files)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // 1 minute for larger files
	defer cancel()

	// Get file info for progress calculation
	fileInfo, _ := os.Stat(videoFile)
	var fileSize int64
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}
	
	log.Printf("Starting FFmpeg processing with %s format (file size: %.2f MB)", 
		audioFormat, float64(fileSize)/(1024*1024))

	var cmd *exec.Cmd
	switch audioFormat {
	case "wav":
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "pcm_s16le", "-ar", "44100", "-y", audioFile)
	case "flac":
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "flac", "-ar", "44100", "-y", audioFile)
	case "aac":
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "aac", "-ab", audioQuality, "-ar", "44100", "-y", audioFile)
	default: // mp3
		cmd = exec.CommandContext(ctx, "ffmpeg", "-i", videoFile, "-vn", "-acodec", "mp3", "-ab", audioQuality, "-ar", "44100", "-y", audioFile)
	}
	
	log.Printf("Processing audio extraction (%.2f MB video → %s audio)...", 
		float64(fileSize)/(1024*1024), audioFormat)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			response := FFmpegResponse{
				Success: false,
				Error:   "FFmpeg processing timed out (60s limit). File may be too large for processing.",
				Progress: "Processing failed - timeout",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
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

	// If using R2 storage, return base64 encoded data
	if req.UseR2Storage {
		// Read the audio file
		audioData, err := os.ReadFile(audioFile)
		if err != nil {
			response := FFmpegResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to read audio file: %v", err),
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Encode to base64
		base64Data := base64.StdEncoding.EncodeToString(audioData)

		// Clean up the local file immediately
		os.Remove(audioFile)

		// Get audio file info
		audioInfo, _ := os.Stat(audioFile)
		var audioSize string
		if audioInfo != nil {
			audioSize = fmt.Sprintf("%.2f MB", float64(audioInfo.Size())/(1024*1024))
		}
		
		response := FFmpegResponse{
			Success:     true,
			Message:     fmt.Sprintf("Audio extracted successfully (%.2f MB video → %s audio)", float64(fileSize)/(1024*1024), audioSize),
			AudioData:   base64Data,
			VideoTitle:  videoTitle,
			VideoSource: videoSource,
			Duration:    duration,
			FileSize:    fmt.Sprintf("%.2f MB", float64(fileSize)/(1024*1024)),
			Progress:    "100%",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get audio file info for the download response
	audioInfo, _ := os.Stat(audioFile)
	var audioSize string
	if audioInfo != nil {
		audioSize = fmt.Sprintf("%.2f MB", float64(audioInfo.Size())/(1024*1024))
	}
	
	// For traditional download, return the download URL
	response := FFmpegResponse{
		Success:     true,
		Message:     fmt.Sprintf("Audio extracted successfully (%.2f MB video → %s audio)", float64(fileSize)/(1024*1024), audioSize),
		AudioURL:    fmt.Sprintf("/download/%s", filepath.Base(audioFile)),
		VideoTitle:  videoTitle,
		VideoSource: videoSource,
		FileSize:    fmt.Sprintf("%.2f MB", float64(fileSize)/(1024*1024)),
		Progress:    "100%",
		Duration:    duration,
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
	router.HandleFunc("/readme", readmeHandler)
	router.HandleFunc("/ffmpeg/extract-audio", ffmpegHandler)
	router.HandleFunc("/ffmpeg/upload", uploadHandler)
	router.HandleFunc("/ffmpeg/upload-base64", uploadBase64Handler)
	router.HandleFunc("/download/", downloadHandler)
	router.HandleFunc("/error", errorHandler)
	router.HandleFunc("/container", handler)
	router.HandleFunc("/", handler)

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
