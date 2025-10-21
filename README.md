# Containers Starter

[![Deploy to Cloudflare](https://deploy.workers.cloudflare.com/button)](https://deploy.workers.cloudflare.com/?url=https://github.com/cloudflare/templates/tree/main/containers-template)

![Containers Template Preview](https://imagedelivery.net/_yJ02hpOMj_EnGvsU2aygw/5aba1fb7-b937-46fd-fa67-138221082200/public)

<!-- dash-content-start -->

This is a [Container](https://developers.cloudflare.com/containers/) starter template.

It demonstrates basic Container configuration, launching and routing to individual container, load balancing over multiple container, running basic hooks on container status changes.

<!-- dash-content-end -->

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
| `POST` | `/ffmpeg/{instance-id}` | Extract audio from video (200MB max, chunked download) | ✅ Active |
| `GET` | `/download/{filename}` | Download extracted audio from R2 storage | ✅ Active |
| `GET` | `/container/{instance-id}` | Direct container access with 2m timeout | ✅ Active |
| `GET` | `/singleton` | Get singleton container instance | ✅ Active |
| `GET` | `/lb` | Load-balanced requests over multiple containers | ✅ Active |
| `GET` | `/error` | Test error handling (development) | ⚠️ Dev only |

### 🎯 How to Extract Audio

#### 1. Basic Usage

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

#### 2. Download the Extracted Audio

```bash
curl "https://vegvisr-container.torarnehave.workers.dev/download/audio_my-audio-job_1760939604006.mp3" \
  -o extracted_audio.mp3
```

#### 3. Real Examples (Tested & Working)

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
```

#### 4. Using JavaScript/Node.js

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

#### 5. Using Python

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

#### 6. Complete Workflow Script

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

#### 3. 📥 Download Extracted Audio

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

#### 4. 🚀 Container Management Endpoints

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

// Usage
const api = new AudioExtractionAPI();

api.extractAndDownload('https://example.com/video.mp4', 'my-job', 'mp3')
  .then(({ metadata, audio }) => {
    console.log('✅ Success:', metadata);
    console.log('🎵 Audio blob size:', audio.size, 'bytes');
    
    // Create download link
    const url = URL.createObjectURL(audio);
    const a = document.createElement('a');
    a.href = url;
    a.download = metadata.file_name || 'audio.mp3';
    a.click();
    URL.revokeObjectURL(url);
  })
  .catch(error => console.error('❌ Error:', error));
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
