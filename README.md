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

## 🎵 FFmpeg Audio Extraction Service

This container-based Cloudflare Worker includes a powerful FFmpeg endpoint for extracting MP3 audio from video files. Perfect for building audio processing applications, podcast creators, or any service that needs to convert video content to audio.

### 🚀 Quick Start

**Your deployed service is available at:**
```
https://vegvisr-container.torarnehave.workers.dev
```

### 📋 Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/` | List all available endpoints |
| `POST` | `/ffmpeg/{instance-id}` | Extract MP3 audio from video |
| `GET` | `/container/{instance-id}` | Access container directly |
| `GET` | `/singleton` | Get singleton container |
| `GET` | `/lb` | Load-balanced container access |

### 🎯 How to Extract Audio

#### 1. Basic Usage

Send a POST request with a video URL:

```bash
curl -X POST "https://vegvisr-container.torarnehave.workers.dev/ffmpeg/my-audio-job" \
  -H "Content-Type: application/json" \
  -d '{"video_url": "https://example.com/video.mp4"}'
```

#### 2. Real Example

```bash
# Extract audio from a video file
curl -X POST "https://vegvisr-container.torarnehave.workers.dev/ffmpeg/demo-instance" \
  -H "Content-Type: application/json" \
  -d '{"video_url": "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"}'
```

#### 3. Using JavaScript/Node.js

```javascript
async function extractAudio(videoUrl, instanceId = 'default') {
  const response = await fetch(`https://vegvisr-container.torarnehave.workers.dev/ffmpeg/${instanceId}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      video_url: videoUrl
    })
  });

  const result = await response.json();
  
  if (result.success) {
    console.log('Audio extracted successfully!');
    console.log('Download URL:', result.audio_url);
    return result;
  } else {
    console.error('Extraction failed:', result.error);
    throw new Error(result.error);
  }
}

// Usage
extractAudio('https://example.com/my-video.mp4')
  .then(result => console.log('Success:', result))
  .catch(error => console.error('Error:', error));
```

#### 4. Using Python

```python
import requests
import json

def extract_audio(video_url, instance_id='default'):
    url = f'https://vegvisr-container.torarnehave.workers.dev/ffmpeg/{instance_id}'
    
    payload = {
        'video_url': video_url
    }
    
    response = requests.post(url, json=payload)
    result = response.json()
    
    if result['success']:
        print('Audio extracted successfully!')
        return result['audio_url']
    else:
        raise Exception(f"Extraction failed: {result['error']}")

# Usage
try:
    audio_url = extract_audio('https://example.com/video.mp4')
    print(f'Audio ready at: {audio_url}')
except Exception as e:
    print(f'Error: {e}')
```

### 📥 Downloading Extracted Audio

Once audio extraction is complete, download the MP3 file:

```bash
# Replace 'audio_filename.mp3' with the actual filename from the response
curl "https://vegvisr-container.torarnehave.workers.dev/container/my-audio-job/download/audio_filename.mp3" \
  -o extracted_audio.mp3
```

#### Complete Workflow Example

```bash
#!/bin/bash

# 1. Extract audio
RESPONSE=$(curl -s -X POST "https://vegvisr-container.torarnehave.workers.dev/ffmpeg/batch-job-1" \
  -H "Content-Type: application/json" \
  -d '{"video_url": "https://example.com/video.mp4"}')

echo "Extraction response: $RESPONSE"

# 2. Parse response and download if successful
if echo "$RESPONSE" | grep -q '"success":true'; then
    AUDIO_URL=$(echo "$RESPONSE" | grep -o '"/download/[^"]*"' | tr -d '"')
    echo "Downloading audio from: $AUDIO_URL"
    
    curl "https://vegvisr-container.torarnehave.workers.dev/container/batch-job-1$AUDIO_URL" \
      -o "extracted_$(date +%s).mp3"
    
    echo "✅ Audio extraction and download complete!"
else
    echo "❌ Audio extraction failed"
    echo "$RESPONSE"
fi
```

### 📝 API Reference

#### Request Format

```json
{
  "video_url": "string (required) - Direct URL to the video file"
}
```

**Supported video formats:** MP4, AVI, MOV, MKV, WebM, FLV, and more (any format FFmpeg supports)

#### Success Response

```json
{
  "success": true,
  "message": "Audio extracted successfully",
  "audio_url": "/download/audio_instanceid_timestamp.mp3"
}
```

#### Error Response

```json
{
  "success": false,
  "error": "Detailed error description"
}
```

#### Common Error Types

- `"video_url is required"` - Missing video URL in request
- `"Failed to download video: HTTP 404"` - Video URL not accessible
- `"FFmpeg failed: ..."` - Video format not supported or corrupted
- `"Failed to create temp file"` - Internal server error

### ⚙️ Technical Details

#### Audio Output Specifications

- **Format:** MP3
- **Bitrate:** 192 kbps
- **Sample Rate:** 44.1 kHz
- **Channels:** Stereo (preserves original channel configuration)

#### Container Instance Management

Each `{instance-id}` creates an isolated container:

- **Parallel Processing:** Different instance IDs process simultaneously
- **Resource Isolation:** Each container has its own temporary storage
- **Automatic Cleanup:** Files are cleaned up after 5 minutes
- **Timeout:** Container sleeps after 2 minutes of inactivity

#### Performance & Limits

- **File Size:** No explicit limit (depends on Cloudflare Worker limits)
- **Processing Time:** Varies by video length and size
- **Concurrent Jobs:** Multiple instances can run in parallel
- **Storage:** Temporary files are automatically cleaned up

### 🛠️ Development & Customization

#### Modify FFmpeg Parameters

Edit `container_src/main.go` to customize the FFmpeg command:

```go
// Current command: Extract MP3 with 192k bitrate
cmd := exec.Command("ffmpeg", "-i", videoFile, "-vn", "-acodec", "mp3", "-ab", "192k", "-ar", "44100", "-y", audioFile)

// Example: Extract as WAV instead
cmd := exec.Command("ffmpeg", "-i", videoFile, "-vn", "-acodec", "pcm_s16le", "-ar", "44100", "-y", audioFile)

// Example: Different quality settings
cmd := exec.Command("ffmpeg", "-i", videoFile, "-vn", "-acodec", "mp3", "-ab", "320k", "-ar", "48000", "-y", audioFile)
```

#### Add New Endpoints

Add new routes in `container_src/main.go`:

```go
router.HandleFunc("/ffmpeg/extract-video-segment", segmentHandler)
router.HandleFunc("/ffmpeg/convert-format", convertHandler)
```

### 🔒 Security Considerations

- **Input Validation:** Only HTTP/HTTPS URLs are accepted
- **File Path Protection:** Download paths are validated to prevent directory traversal
- **Temporary Storage:** Files are stored in isolated `/tmp/processing` directory
- **Auto Cleanup:** All temporary files are automatically removed

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
   - Ensure the video file isn't corrupted

3. **Timeout errors**
   - Large video files may take longer to process
   - Consider splitting large files or using smaller videos

#### Getting Help

- Check the [Cloudflare Containers documentation](https://developers.cloudflare.com/containers/)
- Review the [FFmpeg documentation](https://ffmpeg.org/documentation.html)
- Open an issue in this repository for bugs or feature requests

### 📈 Production Usage

For production deployments:

1. **Custom Domain:** Set up a custom domain instead of using `.workers.dev`
2. **Rate Limiting:** Implement rate limiting for your use case
3. **Authentication:** Add API key authentication if needed
4. **Monitoring:** Set up logging and monitoring for your endpoints
5. **CDN:** Consider using Cloudflare's CDN for serving downloaded files

Your feedback and contributions are welcome!
