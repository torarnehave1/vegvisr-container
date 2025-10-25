# 🎵 Vegvisr Container - Video-to-Audio Conversion Service

[![Deploy to Cloudflare](https://deploy.workers.cloudflare.com/button)](https://deploy.workers.cloudflare.com/?url=https://github.com/cloudflare/templates/tree/main/containers-template)

![Containers Template Preview](https://imagedelivery.net/_yJ02hpOMj_EnGvsU2aygw/5aba1fb7-b937-46fd-fa67-138221082200/public)

## 🎯 What This Service Does

**Vegvisr Container** is a powerful **Cloudflare Workers-based video-to-audio conversion service** with cloud storage capabilities. It provides a robust, scalable API for extracting high-quality audio from video files using FFmpeg and storing them in Cloudflare R2 cloud storage.

### 🎵 Core Functionality

**Primary Purpose:** Extract audio (MP3, WAV, AAC, FLAC) from video files using FFmpeg and store them in Cloudflare R2 cloud storage with global CDN distribution.

### 🚀 Key Features

- **🔗 URL Processing:** Download videos from any public URL and extract audio
- **📤 Direct Upload:** Accept video file uploads directly from frontend applications  
- **☁️ Cloud Storage:** Automatic storage in Cloudflare R2 with global distribution
- **⚡ High Performance:** Chunked downloading supports files up to 200MB
- **🎛️ Multiple Formats:** Output as MP3, WAV, AAC, or FLAC
- **📊 Progress Tracking:** Real-time file size detection and processing updates
- **🌍 Global Access:** Worldwide availability via Cloudflare's edge network
- **🔄 Scalable:** Multiple container instances for concurrent processing
- **📚 Self-Documenting:** Built-in API documentation and examples

### 💼 Use Cases

- **🎙️ Content Creation:** Convert video content to audio for podcasts and media production
- **🎬 Media Processing:** Batch convert video libraries to audio formats
- **🌐 Web Applications:** Backend service for video-to-audio conversion in web apps
- **📱 Social Media:** Extract audio from video posts and content
- **🎓 Education:** Convert educational videos to audio format for accessibility
- **♿ Accessibility:** Provide audio alternatives for video content
- **🏢 Enterprise:** Integrate into content management systems and workflows

### 🏗️ Architecture Overview

**Technology Stack:**
- **⚡ Runtime:** Cloudflare Workers with Containers (Durable Objects)
- **🐧 Container:** Alpine Linux with FFmpeg v6.1.2
- **🔧 Backend:** Go HTTP server for high-performance video processing
- **🌐 Frontend:** TypeScript/Hono for routing and R2 integration
- **💾 Storage:** Cloudflare R2 bucket with S3-compatible API
- **🔒 Security:** Container isolation and input validation

**Processing Flow:**
1. **📥 Input:** Accept video via URL or direct file upload
2. **⬇️ Download:** Chunked downloading with progress tracking (5MB chunks)
3. **🎵 Extract:** FFmpeg audio extraction with configurable quality settings
4. **☁️ Upload:** Automatic upload to Cloudflare R2 storage
5. **🔗 Access:** Provide global download URL for immediate access

### 📊 Technical Specifications

- **📏 File Size Limit:** 200MB with chunked downloading
- **⏱️ Processing Timeout:** 60 seconds for FFmpeg processing
- **🏃 Container Timeout:** 2 minutes before auto-sleep (restart on demand)
- **🔄 Concurrency:** Unlimited parallel processing with unique instance IDs
- **🌐 Global Distribution:** Cloudflare's 200+ edge locations
- **📈 Reliability:** Automatic error handling and recovery

Outside of this repo, you can start a new project with this template using [C3](https://developers.cloudflare.com/pages/get-started/c3/) (the `create-cloudflare` CLI):

```bash
npm create cloudflare@latest -- --template=cloudflare/templates/containers-template
```

## Getting Started

First, run:

```bash
npm install
# or
yarn install
# or
pnpm install
# or
bun install
```

Then run the development server (using the package manager of your choice):

```bash
npm run dev
```

Open [http://localhost:8787](http://localhost:8787) with your browser to see the result.

You can start editing your Worker by modifying `src/index.ts` and you can start
editing your Container by editing the content of `container_src`.

## Deploying To Production

| Command          | Action                                |
| :--------------- | :------------------------------------ |
| `npm run deploy` | Deploy your application to Cloudflare |

## Learn More

To learn more about Containers, take a look at the following resources:

- [Container Documentation](https://developers.cloudflare.com/containers/) - learn about Containers
- [Container Class](https://github.com/cloudflare/containers) - learn about the Container helper class

## 🎵 FFmpeg Audio Extraction Service with Cloudflare R2 Storage

This container-based Cloudflare Worker provides a powerful FFmpeg endpoint for extracting MP3 audio from video files with automatic cloud storage. Perfect for building audio processing applications, podcast creators, content management systems, or any service that needs to convert video content to audio.

### 🚀 Quick Start

**Your deployed service is available at:**
```
https://vegvisr-container.torarnehave.workers.dev
```

### 🗄️ Storage Architecture

- **R2 Bucket**: `audio-from-video-files`
- **Storage Binding**: `AUDIO_STORAGE`
- **File Naming**: `audio_{instance-id}_{timestamp}.mp3`
- **Auto-cleanup**: Files are managed efficiently with R2's durable storage
- **Global Access**: Available worldwide via Cloudflare's edge network

### 📋 API Endpoints

| Method | Endpoint | Description | Status |
|--------|----------|-------------|---------|
| `GET` | `/` | List all available endpoints | ✅ Active |
| `POST` | `/ffmpeg/{instance-id}` | Extract audio from video URL (200MB max, chunked download) | ✅ Active |
| `POST` | `/upload/{instance-id}` | Upload video files directly (auto-handles small & large files up to 100MB) | ✅ Active |
| `GET` | `/download/{filename}` | Download extracted audio from R2 storage | ✅ Active |
| `GET` | `/readme` | Download README.md documentation from GitHub | ✅ Active |
| `GET` | `/container/{instance-id}` | Direct container access with 2m timeout | ✅ Active |
| `GET` | `/singleton` | Get singleton container instance | ✅ Active |
| `GET` | `/lb` | Load-balanced requests over multiple containers | ✅ Active |
| `GET` | `/error` | Test error handling (development) | ⚠️ Dev only |

### 🎯 How to Extract Audio

#### 1. Basic Usage (URL-based)

Send a POST request with a video URL:

```bash
curl -X POST "https://vegvisr-container.torarnehave.workers.dev/ffmpeg/my-audio-job" \
  -H "Content-Type: application/json" \
  -d '{"video_url": "https://example.com/video.mp4"}'
```

**Response:**
```json
{
  "success": true,
  "message": "Audio extracted and stored successfully",
  "download_url": "/download/audio_my-audio-job_1760939604006.mp3",
  "file_name": "audio_my-audio-job_1760939604006.mp3",
  "r2_key": "audio_my-audio-job_1760939604006.mp3"
}
```

#### 2. Upload Files Directly (All Sizes - Smart Auto-Detection)

**ONE endpoint handles everything - from screenshots to 10-minute presentations!**

```bash
# Small files (< 15MB) - uses direct processing
curl -X POST \
  -F "video=@screenshot.mp4" \
  -F "output_format=mp3" \
  https://vegvisr-container.torarnehave.workers.dev/upload/my-job

# Large files (15-100MB) - automatically uploads to R2 first
curl -X POST \
  -F "video=@presentation.mov" \
  -F "output_format=mp3" \
  https://vegvisr-container.torarnehave.workers.dev/upload/my-presentation \
  --max-time 120
```

**How it works (automatically):**
- **Files ≤ 15MB**: Direct processing (fast, ~5-15 seconds)
- **Files > 15MB**: Uploads to R2 first, then processes (handles up to 100MB, ~30-60 seconds)

**Perfect for:**
- ✅ Screenshots and short clips
- ✅ QuickTime screen recordings (.mov)
- ✅ 10-minute presentations (tested with 26MB .mov file)
- ✅ File sizes up to 100MB

**Example Response:**
```json
{
  "success": true,
  "message": "Audio extracted from uploaded file successfully",
  "download_url": "/download/audio_my-presentation_1761395200609.mp3",
  "file_name": "audio_my-presentation_1761395200609.mp3",
  "r2_key": "audio_my-presentation_1761395200609.mp3",
  "audio_url": "/download/audio_my-presentation_1761395200609.mp3"
}
```

#### 3. Download the Extracted Audio

```bash
curl "https://vegvisr-container.torarnehave.workers.dev/download/audio_my-audio-job_1760939604006.mp3" \
  -o extracted_audio.mp3
```

#### 4. Real Examples (Tested & Working)

```bash
# Example 1: Extract from 13MB video (ForBiggerBlazes.mp4)
curl -X POST "https://vegvisr-container.torarnehave.workers.dev/ffmpeg/test-large-1" \
  -H "Content-Type: application/json" \
  -d '{"video_url": "https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4", "output_format": "mp3"}'

# Response: {"success":true,"message":"Audio extracted and stored successfully","download_url":"/download/audio_test-large-1_1761032972174.mp3"...}

# Example 2: Extract from 16MB video (ForBiggerEscapes.mp4) 
curl -X POST "https://vegvisr-container.torarnehave.workers.dev/ffmpeg/test-large-2" \
  -H "Content-Type: application/json" \
  -d '{"video_url": "https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4", "output_format": "mp3"}'

# Both examples demonstrate chunked downloading handling files larger than 5MB

# Example 3: Upload 26MB QuickTime screen recording (tested successfully)
# ONE unified endpoint automatically handles large files!
curl -X POST \
  -F "video=@screen-recording.mov" \
  -F "output_format=mp3" \
  https://vegvisr-container.torarnehave.workers.dev/upload/my-presentation \
  --max-time 120

# Response: {"success":true,"message":"Audio extracted from uploaded file successfully"...}
# The endpoint automatically detected it's a large file and uploaded to R2 first!
```

#### 5. Using JavaScript/Node.js

```javascript
async function extractAndDownloadAudio(videoUrl, instanceId = 'default') {
  // Step 1: Extract audio
  const extractResponse = await fetch(`https://vegvisr-container.torarnehave.workers.dev/ffmpeg/${instanceId}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      video_url: videoUrl
    })
  });

  const result = await extractResponse.json();
  
  if (!result.success) {
    throw new Error(`Extraction failed: ${result.error}`);
  }

  console.log('✅ Audio extracted successfully!');
  console.log('📁 File stored in R2:', result.file_name);
  
  // Step 2: Download the audio file
  const downloadResponse = await fetch(`https://vegvisr-container.torarnehave.workers.dev${result.download_url}`);
  
  if (!downloadResponse.ok) {
    throw new Error('Download failed');
  }

  // Return as blob for browser use or buffer for Node.js
  return await downloadResponse.blob(); // or .arrayBuffer() for Node.js
}

// Usage
extractAndDownloadAudio('https://example.com/my-video.mp4', 'my-job')
  .then(audioBlob => {
    console.log('🎵 Audio ready!', audioBlob);
    // Create download link, save to file, etc.
  })
  .catch(error => console.error('❌ Error:', error));
```

#### 6. Using Python

```python
import requests
import json

def extract_and_download_audio(video_url, instance_id='default'):
    # Step 1: Extract audio
    extract_url = f'https://vegvisr-container.torarnehave.workers.dev/ffmpeg/{instance_id}'
    
    payload = {'video_url': video_url}
    response = requests.post(extract_url, json=payload)
    result = response.json()
    
    if not result['success']:
        raise Exception(f"Extraction failed: {result['error']}")
    
    print(f"✅ Audio extracted: {result['file_name']}")
    
    # Step 2: Download the audio
    download_url = f"https://vegvisr-container.torarnehave.workers.dev{result['download_url']}"
    audio_response = requests.get(download_url)
    
    if audio_response.status_code != 200:
        raise Exception('Download failed')
    
    # Save to file
    filename = result['file_name']
    with open(filename, 'wb') as f:
        f.write(audio_response.content)
    
    print(f"🎵 Audio saved as: {filename}")
    return filename

# Usage
try:
    audio_file = extract_and_download_audio('https://example.com/video.mp4', 'python-job')
    print(f'Success! Audio saved: {audio_file}')
except Exception as e:
    print(f'Error: {e}')
```

#### 7. Complete Workflow Script

```bash
#!/bin/bash

# Complete workflow with error handling
VIDEO_URL="https://www.learningcontainer.com/wp-content/uploads/2020/05/sample-mp4-file.mp4"
INSTANCE_ID="batch-job-$(date +%s)"
WORKER_URL="https://vegvisr-container.torarnehave.workers.dev"

echo "🎬 Extracting audio from: $VIDEO_URL"
echo "📋 Instance ID: $INSTANCE_ID"

# Step 1: Extract audio
RESPONSE=$(curl -s -X POST "$WORKER_URL/ffmpeg/$INSTANCE_ID" \
  -H "Content-Type: application/json" \
  -d "{\"video_url\": \"$VIDEO_URL\"}")

echo "📄 Extraction response: $RESPONSE"

# Check if extraction was successful
if echo "$RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    echo "✅ Audio extraction successful!"
    
    # Extract download URL
    DOWNLOAD_URL=$(echo "$RESPONSE" | jq -r '.download_url')
    FILENAME=$(echo "$RESPONSE" | jq -r '.file_name')
    
    echo "📥 Downloading: $FILENAME"
    
    # Step 2: Download the audio
    curl -s "$WORKER_URL$DOWNLOAD_URL" -o "$FILENAME"
    
    if [ $? -eq 0 ]; then
        echo "🎵 Success! Audio saved as: $FILENAME"
        echo "📊 File info:"
        file "$FILENAME"
        ls -lh "$FILENAME"
    else
        echo "❌ Download failed"
        exit 1
    fi
else
    echo "❌ Audio extraction failed"
    echo "$RESPONSE" | jq -r '.error // "Unknown error"'
    exit 1
fi
```

### 📝 Complete API Reference

#### 1. 📋 Root Endpoint - Service Discovery

**Endpoint:** `GET /`

**Description:** Returns all available endpoints with current capabilities

**Response:**
```text
Available endpoints:
GET /container/<ID> - Start a container for each ID with a 2m timeout
GET /lb - Load balance requests over multiple containers
GET /error - Start a container that errors (demonstrates error handling)
GET /singleton - Get a single specific container instance
POST /ffmpeg/<ID> - Extract MP3 audio from video using FFmpeg
  • Supports files up to 200MB with chunked downloading
  • Progress tracking and detailed file size information
  • Automatic R2 storage for extracted audio
GET /download/<filename> - Download extracted audio files from R2 storage
```

#### 2. 🎵 FFmpeg Audio Extraction

**Endpoint:** `POST /ffmpeg/{instance-id}`

**Description:** Extract audio from video files with chunked downloading support up to 200MB

**Parameters:**
- `instance-id` (path): Unique identifier for the processing job (URL-safe string)

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "video_url": "string (required) - Direct HTTP/HTTPS URL to video file",
  "output_format": "string (optional) - Audio format: mp3, aac, wav, flac (default: mp3)"
}
```

**Supported Input Formats:** MP4, AVI, MOV, MKV, WebM, FLV, 3GP, WMV, and all FFmpeg-supported formats

**File Size Limits:**
- Maximum: 200MB (with chunked downloading in 5MB segments)
- Progress tracking: Real-time file size detection and download progress
- Timeout: 60 seconds for FFmpeg processing

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Audio extracted and stored successfully",
  "download_url": "/download/audio_instanceid_timestamp.mp3",
  "file_name": "audio_instanceid_timestamp.mp3", 
  "r2_key": "audio_instanceid_timestamp.mp3",
  "audio_url": "/download/audio_instanceid_timestamp.mp3"
}
```

**Error Responses:**

**400 Bad Request - Missing video_url:**
```json
{
  "success": false,
  "message": "",
  "error": "video_url is required"
}
```

**400 Bad Request - File too large:**
```json
{
  "success": false,
  "message": "",
  "error": "video file too large (250.5 MB). Maximum supported: 200MB"
}
```

**500 Internal Server Error - Download failed:**
```json
{
  "success": false, 
  "message": "",
  "error": "failed to get file info: Head \"https://example.com/video.mp4\": context deadline exceeded"
}
```

**500 Internal Server Error - FFmpeg processing failed:**
```json
{
  "success": false,
  "message": "",
  "error": "ffmpeg failed: exit status 1, stderr: [error details]"
}
```

#### 3. � Direct File Upload

**Endpoint:** `POST /upload/{instance-id}`

**Description:** Upload video files directly from frontend applications for audio extraction

**Parameters:**
- `instance-id` (path): Unique identifier for the processing job (URL-safe string)

**Request Headers:**
```
Content-Type: multipart/form-data
```

**Request Body (multipart/form-data):**
```
video: [File] (required) - Video file to process
output_format: string (optional) - Audio format: mp3, aac, wav, flac (default: mp3)
instance_id: string (optional) - Override instance ID from URL path
```

**File Size Limits:**
- Maximum: 200MB per upload
- Supported formats: MP4, AVI, MOV, MKV, WebM, FLV, 3GP, WMV, and all FFmpeg-supported formats
- Processing timeout: 60 seconds for FFmpeg

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Audio extracted from uploaded file and stored successfully",
  "download_url": "/download/audio_instanceid_timestamp.mp3",
  "file_name": "audio_instanceid_timestamp.mp3", 
  "r2_key": "audio_instanceid_timestamp.mp3",
  "audio_url": "/download/audio_instanceid_timestamp.mp3"
}
```

**Error Responses:**

**400 Bad Request - No file uploaded:**
```json
{
  "success": false,
  "error": "Failed to get uploaded file: http: no such file"
}
```

**400 Bad Request - File too large:**
```json
{
  "success": false,
  "error": "File too large (250.5 MB). Maximum supported: 200MB"
}
```

**400 Bad Request - Unsupported format:**
```json
{
  "success": false,
  "error": "Unsupported output format: xyz. Supported: mp3, wav, aac, flac"
}
```

**500 Internal Server Error - Processing failed:**
```json
{
  "success": false,
  "error": "ffmpeg failed: exit status 1, stderr: [error details]"
}
```

#### 4. �📥 Download Extracted Audio

**Endpoint:** `GET /download/{filename}`

**Description:** Download processed audio files from Cloudflare R2 storage

**Parameters:**
- `filename` (path): Audio filename returned from FFmpeg endpoint (format: `audio_{instance-id}_{timestamp}.{format}`)

**Response Headers:**
```
Content-Type: audio/mpeg (or appropriate MIME type)
Content-Disposition: attachment; filename="{filename}"
```

**Success Response (200 OK):** Binary audio file content

**Error Response (404 Not Found):** File not found in R2 storage

#### 5. 🚀 Container Management Endpoints

**Direct Container Access:** `GET /container/{instance-id}`
- Creates or connects to specific container instance
- 2-minute timeout before container sleeps
- Returns container connection info

**Load Balanced Access:** `GET /lb`  
- Distributes requests across multiple containers
- Automatic scaling and load distribution
- Returns balanced container response

**Singleton Container:** `GET /singleton`
- Access to dedicated singleton container instance
- Persistent for debugging and development
- Returns singleton container info

**Error Testing:** `GET /error`
- Triggers container error for testing error handling
- Development/debugging endpoint only
- Returns error demonstration

### 🔧 Integration Examples

#### cURL Examples

**Basic audio extraction:**
```bash
curl -X POST "https://vegvisr-container.torarnehave.workers.dev/ffmpeg/job123" \
  -H "Content-Type: application/json" \
  -d '{"video_url": "https://example.com/video.mp4"}'
```

**Extract specific format:**
```bash
curl -X POST "https://vegvisr-container.torarnehave.workers.dev/ffmpeg/job123" \
  -H "Content-Type: application/json" \
  -d '{"video_url": "https://example.com/video.mp4", "output_format": "wav"}'
```

**Download extracted audio:**
```bash
curl "https://vegvisr-container.torarnehave.workers.dev/download/audio_job123_1761032972174.mp3" \
  -o extracted_audio.mp3
```

**Upload video file directly:**
```bash
curl -X POST "https://vegvisr-container.torarnehave.workers.dev/upload/upload-job-123" \
  -F "video=@/path/to/your/video.mp4" \
  -F "output_format=mp3"
```

**Upload with specific format:**
```bash
curl -X POST "https://vegvisr-container.torarnehave.workers.dev/upload/upload-job-456" \
  -F "video=@/path/to/your/video.mov" \
  -F "output_format=wav"
```

#### HTML File Upload Form

```html
<!DOCTYPE html>
<html>
<head>
    <title>Video to Audio Converter</title>
</head>
<body>
    <h2>🎵 Upload Video for Audio Extraction</h2>
    
    <form id="uploadForm" enctype="multipart/form-data">
        <div>
            <label for="videoFile">Select Video File (Max 200MB):</label><br>
            <input type="file" id="videoFile" name="video" accept="video/*" required>
        </div>
        
        <div>
            <label for="format">Output Format:</label><br>
            <select id="format" name="output_format">
                <option value="mp3">MP3 (Default)</option>
                <option value="wav">WAV (High Quality)</option>
                <option value="aac">AAC (Compressed)</option>
                <option value="flac">FLAC (Lossless)</option>
            </select>
        </div>
        
        <div>
            <button type="submit">🎵 Extract Audio</button>
        </div>
    </form>
    
    <div id="progress" style="display: none;">
        <p>⏳ Processing your video...</p>
        <progress id="progressBar" style="width: 100%;"></progress>
    </div>
    
    <div id="result" style="display: none;">
        <h3>✅ Audio Extracted Successfully!</h3>
        <p id="resultMessage"></p>
        <a id="downloadLink" href="#" download>📥 Download Audio File</a>
    </div>
    
    <div id="error" style="display: none; color: red;">
        <h3>❌ Error</h3>
        <p id="errorMessage"></p>
    </div>

    <script>
        document.getElementById('uploadForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const fileInput = document.getElementById('videoFile');
            const formatSelect = document.getElementById('format');
            const progressDiv = document.getElementById('progress');
            const resultDiv = document.getElementById('result');
            const errorDiv = document.getElementById('error');
            
            if (!fileInput.files[0]) {
                alert('Please select a video file');
                return;
            }
            
            // Check file size (200MB limit)
            const file = fileInput.files[0];
            const maxSize = 200 * 1024 * 1024; // 200MB
            if (file.size > maxSize) {
                alert(`File too large (${(file.size / (1024*1024)).toFixed(1)}MB). Maximum: 200MB`);
                return;
            }
            
            // Hide previous results
            resultDiv.style.display = 'none';
            errorDiv.style.display = 'none';
            progressDiv.style.display = 'block';
            
            try {
                const formData = new FormData();
                formData.append('video', file);
                formData.append('output_format', formatSelect.value);
                
                const instanceId = 'upload-' + Date.now();
                const response = await fetch(`https://vegvisr-container.torarnehave.workers.dev/upload/${instanceId}`, {
                    method: 'POST',
                    body: formData
                });
                
                const result = await response.json();
                
                progressDiv.style.display = 'none';
                
                if (result.success) {
                    document.getElementById('resultMessage').textContent = 
                        `File processed: ${result.file_name}`;
                    document.getElementById('downloadLink').href = 
                        `https://vegvisr-container.torarnehave.workers.dev${result.download_url}`;
                    document.getElementById('downloadLink').download = result.file_name;
                    resultDiv.style.display = 'block';
                } else {
                    document.getElementById('errorMessage').textContent = result.error;
                    errorDiv.style.display = 'block';
                }
            } catch (error) {
                progressDiv.style.display = 'none';
                document.getElementById('errorMessage').textContent = 
                    'Network error: ' + error.message;
                errorDiv.style.display = 'block';
            }
        });
    </script>
</body>
</html>
```

#### JavaScript/TypeScript Integration

```typescript
interface FFmpegRequest {
  video_url: string;
  output_format?: 'mp3' | 'aac' | 'wav' | 'flac';
}

interface FFmpegResponse {
  success: boolean;
  message: string;
  download_url?: string;
  file_name?: string;
  r2_key?: string;
  audio_url?: string;
  error?: string;
}

class AudioExtractionAPI {
  private baseUrl = 'https://vegvisr-container.torarnehave.workers.dev';

  async extractAudio(
    videoUrl: string, 
    instanceId: string = `job-${Date.now()}`,
    format: 'mp3' | 'aac' | 'wav' | 'flac' = 'mp3'
  ): Promise<FFmpegResponse> {
    const response = await fetch(`${this.baseUrl}/ffmpeg/${instanceId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        video_url: videoUrl,
        output_format: format
      })
    });

    return await response.json();
  }

  async downloadAudio(filename: string): Promise<Blob> {
    const response = await fetch(`${this.baseUrl}/download/${filename}`);
    
    if (!response.ok) {
      throw new Error(`Download failed: ${response.status}`);
    }
    
    return await response.blob();
  }

  async uploadVideo(
    videoFile: File,
    instanceId: string = `upload-${Date.now()}`,
    format: 'mp3' | 'aac' | 'wav' | 'flac' = 'mp3'
  ): Promise<FFmpegResponse> {
    // Check file size (200MB limit)
    const maxSize = 200 * 1024 * 1024; // 200MB
    if (videoFile.size > maxSize) {
      throw new Error(`File too large (${(videoFile.size / (1024*1024)).toFixed(1)}MB). Maximum: 200MB`);
    }

    const formData = new FormData();
    formData.append('video', videoFile);
    formData.append('output_format', format);

    const response = await fetch(`${this.baseUrl}/upload/${instanceId}`, {
      method: 'POST',
      body: formData
    });

    return await response.json();
  }

  async uploadAndDownload(
    videoFile: File,
    instanceId?: string,
    format?: 'mp3' | 'aac' | 'wav' | 'flac'
  ): Promise<{ metadata: FFmpegResponse; audio: Blob }> {
    const result = await this.uploadVideo(videoFile, instanceId, format);
    
    if (!result.success || !result.file_name) {
      throw new Error(result.error || 'Upload processing failed');
    }
    
    const audio = await this.downloadAudio(result.file_name);
    
    return { metadata: result, audio };
  }

  async extractAndDownload(
    videoUrl: string, 
    instanceId?: string, 
    format?: 'mp3' | 'aac' | 'wav' | 'flac'
  ): Promise<{ metadata: FFmpegResponse; audio: Blob }> {
    const result = await this.extractAudio(videoUrl, instanceId, format);
    
    if (!result.success || !result.file_name) {
      throw new Error(result.error || 'Extraction failed');
    }
    
    const audio = await this.downloadAudio(result.file_name);
    
    return { metadata: result, audio };
  }
}

// Usage Examples
const api = new AudioExtractionAPI();

// Extract from URL
api.extractAndDownload('https://example.com/video.mp4', 'my-job', 'mp3')
  .then(({ metadata, audio }) => {
    console.log('✅ URL Processing Success:', metadata);
    console.log('🎵 Audio blob size:', audio.size, 'bytes');
    
    // Create download link
    const url = URL.createObjectURL(audio);
    const a = document.createElement('a');
    a.href = url;
    a.download = metadata.file_name || 'audio.mp3';
    a.click();
    URL.revokeObjectURL(url);
  })
  .catch(error => console.error('❌ URL Processing Error:', error));

// Upload file directly (from file input)
const fileInput = document.getElementById('videoFile') as HTMLInputElement;
if (fileInput.files && fileInput.files[0]) {
  const videoFile = fileInput.files[0];
  
  api.uploadAndDownload(videoFile, 'upload-job-123', 'wav')
    .then(({ metadata, audio }) => {
      console.log('✅ Upload Processing Success:', metadata);
      console.log('🎵 Audio blob size:', audio.size, 'bytes');
      
      // Create download link  
      const url = URL.createObjectURL(audio);
      const a = document.createElement('a');
      a.href = url;
      a.download = metadata.file_name || 'audio.wav';
      a.click();
      URL.revokeObjectURL(url);
    })
    .catch(error => console.error('❌ Upload Processing Error:', error));
}

// Advanced: Upload with progress tracking
async function uploadWithProgress(file: File) {
  try {
    console.log(`📤 Uploading ${file.name} (${(file.size / (1024*1024)).toFixed(1)}MB)`);
    
    const result = await api.uploadVideo(file, `upload-${Date.now()}`, 'mp3');
    
    if (result.success) {
      console.log('🎵 Processing complete!');
      const audio = await api.downloadAudio(result.file_name!);
      console.log(`📥 Downloaded ${audio.size} bytes`);
      return audio;
    } else {
      throw new Error(result.error);
    }
  } catch (error) {
    console.error('❌ Upload failed:', error);
    throw error;
  }
}
```

#### Python Integration

```python
import requests
import json
from typing import Optional, Literal
from dataclasses import dataclass

@dataclass 
class FFmpegResponse:
    success: bool
    message: str
    download_url: Optional[str] = None
    file_name: Optional[str] = None
    r2_key: Optional[str] = None  
    audio_url: Optional[str] = None
    error: Optional[str] = None

class AudioExtractionAPI:
    def __init__(self, base_url: str = 'https://vegvisr-container.torarnehave.workers.dev'):
        self.base_url = base_url

    def extract_audio(
        self, 
        video_url: str, 
        instance_id: Optional[str] = None,
        output_format: Literal['mp3', 'aac', 'wav', 'flac'] = 'mp3'
    ) -> FFmpegResponse:
        """Extract audio from video URL"""
        if instance_id is None:
            import time
            instance_id = f"python-job-{int(time.time())}"
            
        url = f"{self.base_url}/ffmpeg/{instance_id}"
        payload = {
            'video_url': video_url,
            'output_format': output_format
        }
        
        response = requests.post(url, json=payload, timeout=120)
        data = response.json()
        
        return FFmpegResponse(**data)

    def download_audio(self, filename: str) -> bytes:
        """Download audio file as bytes"""
        url = f"{self.base_url}/download/{filename}"
        response = requests.get(url, timeout=60)
        response.raise_for_status()
        return response.content

    def extract_and_save(
        self, 
        video_url: str, 
        output_path: Optional[str] = None,
        instance_id: Optional[str] = None,
        output_format: Literal['mp3', 'aac', 'wav', 'flac'] = 'mp3'
    ) -> str:
        """Extract audio and save to file"""
        # Extract audio
        result = self.extract_audio(video_url, instance_id, output_format)
        
        if not result.success:
            raise Exception(f"Extraction failed: {result.error}")
        
        # Download audio
        audio_data = self.download_audio(result.file_name)
        
        # Save to file
        if output_path is None:
            output_path = result.file_name
            
        with open(output_path, 'wb') as f:
            f.write(audio_data)
            
        return output_path

# Usage examples
api = AudioExtractionAPI()

try:
    # Simple extraction and save
    saved_file = api.extract_and_save(
        'https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4',
        'my_audio.mp3',
        'python-demo'
    )
    print(f"✅ Audio saved to: {saved_file}")
    
    # Advanced usage with error handling
    result = api.extract_audio(
        'https://example.com/video.mp4', 
        'custom-job-123', 
        'wav'
    )
    
    if result.success:
        print(f"📁 File available: {result.download_url}")
        audio_bytes = api.download_audio(result.file_name)
        print(f"🎵 Downloaded {len(audio_bytes)} bytes")
    else:
        print(f"❌ Failed: {result.error}")
        
except Exception as e:
    print(f"Error: {e}")
```

### ⚙️ Technical Details

#### Audio Output Specifications

- **Format:** MP3 (MPEG ADTS, layer III, v1)
- **Bitrate:** 192 kbps
- **Sample Rate:** 44.1 kHz
- **Channels:** Stereo (preserves original channel configuration)
- **Metadata:** ID3 version 2.4.0 tags included

#### Container Instance Management

Each `{instance-id}` creates an isolated container:

- **Parallel Processing:** Different instance IDs process simultaneously
- **Resource Isolation:** Each container has its own temporary processing space
- **Automatic R2 Upload:** Files automatically uploaded to cloud storage
- **Timeout:** Container sleeps after 2 minutes of inactivity
- **Cleanup:** Temporary files removed after R2 upload

#### Cloudflare R2 Storage Benefits

- **Cost-Effective:** No egress fees for downloads
- **Global CDN:** Fast access worldwide via Cloudflare's edge network
- **Durability:** 99.999999999% (11 9's) object durability
- **Scalability:** Handle thousands of concurrent audio extractions
- **S3-Compatible:** Standard APIs for easy integration

#### Performance & Limits

- **File Size Limit:** 200MB maximum with chunked downloading (5MB chunks)
- **Processing Timeout:** 60 seconds for FFmpeg processing
- **Container Timeout:** 2 minutes before container sleeps (auto-restart on new request)
- **Concurrent Jobs:** Unlimited parallel processing with different instance IDs
- **Download Speed:** Chunked downloading prevents timeouts for large files
- **Storage:** Persistent in Cloudflare R2 with global CDN distribution
- **Bandwidth:** Leverages Cloudflare's global network for optimized performance
- **Supported Formats:** All FFmpeg-supported video inputs, multiple audio output formats
- **Progress Tracking:** Real-time file size detection and download progress logging

### 🛠️ Development & Customization

#### Modify FFmpeg Parameters

Edit `container_src/main.go` to customize the FFmpeg command:

```go
// Current command: Extract MP3 with 192k bitrate
cmd := exec.Command("ffmpeg", "-i", videoFile, "-vn", "-acodec", "mp3", "-ab", "192k", "-ar", "44100", "-y", audioFile)

// Example: Extract as WAV instead
cmd := exec.Command("ffmpeg", "-i", videoFile, "-vn", "-acodec", "pcm_s16le", "-ar", "44100", "-y", audioFile)

// Example: Higher quality MP3
cmd := exec.Command("ffmpeg", "-i", videoFile, "-vn", "-acodec", "mp3", "-ab", "320k", "-ar", "48000", "-y", audioFile)

// Example: Extract specific time range (first 30 seconds)
cmd := exec.Command("ffmpeg", "-i", videoFile, "-vn", "-acodec", "mp3", "-ab", "192k", "-ar", "44100", "-t", "30", "-y", audioFile)
```

#### Add New Endpoints

Add new routes in `container_src/main.go`:

```go
router.HandleFunc("/ffmpeg/extract-segment", segmentHandler)
router.HandleFunc("/ffmpeg/convert-format", convertHandler)
router.HandleFunc("/ffmpeg/get-metadata", metadataHandler)
```

#### Configure R2 Storage

The R2 bucket configuration is in `wrangler.jsonc`:

```jsonc
{
  "r2_buckets": [
    {
      "binding": "AUDIO_STORAGE",
      "bucket_name": "audio-from-video-files",
      "remote": true
    }
  ]
}
```

### 🔒 Security Considerations

- **Input Validation:** Only HTTP/HTTPS URLs are accepted
- **File Path Protection:** Download paths are validated to prevent directory traversal
- **Container Isolation:** Each instance runs in its own isolated environment
- **R2 Access Control:** Files are accessible only through your worker endpoints
- **Temporary Processing:** Local files are cleaned up after R2 upload

### 🐛 Troubleshooting

#### Common Issues

1. **"Docker CLI could not be launched"**
   ```bash
   # Install Docker first
   brew install --cask docker
   # Or use Colima
   brew install colima docker && colima start
   ```

2. **"FFmpeg failed" errors**
   - Check if the video URL is accessible
   - Verify the video format is supported
   - Ensure the video file isn't corrupted or protected

3. **"R2 storage failed" errors**
   - Check R2 bucket exists: `wrangler r2 bucket list`
   - Verify bucket permissions in Cloudflare dashboard
   - Ensure R2 binding is configured correctly

4. **Download timeouts**
   - Large video files may take longer to process
   - Consider implementing progress tracking for long operations

#### Getting Help

- Check the [Cloudflare Containers documentation](https://developers.cloudflare.com/containers/)
- Review the [Cloudflare R2 documentation](https://developers.cloudflare.com/r2/)
- Review the [FFmpeg documentation](https://ffmpeg.org/documentation.html)
- Open an issue in this repository for bugs or feature requests

### 📈 Production Usage

For production deployments:

1. **Custom Domain:** Set up a custom domain instead of using `.workers.dev`
2. **Rate Limiting:** Implement rate limiting for your use case
3. **Authentication:** Add API key authentication if needed
4. **Monitoring:** Set up logging and monitoring for your endpoints
5. **R2 Lifecycle:** Configure automatic file cleanup policies
6. **Webhooks:** Add completion notifications for long-running jobs
7. **Queue System:** Implement job queuing for high-volume processing

### 💡 Use Cases

- **Podcast Creation:** Extract audio from video content
- **Content Management:** Automated audio extraction for CMS systems
- **Social Media:** Convert video posts to audio format
- **Education:** Extract audio from educational video content
- **Accessibility:** Provide audio alternatives for video content
- **Batch Processing:** Convert large video libraries to audio

Your feedback and contributions are welcome!
