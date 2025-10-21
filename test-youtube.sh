#!/bin/bash

echo "🧪 Testing yt-dlp and FFmpeg in the deployed container"
echo "=================================================="

# Test with a short public domain video
VIDEO_URL="https://www.youtube.com/watch?v=C0DPdy98e4c"  # Short test video
INSTANCE_ID="yt-dlp-test-$(date +%s)"
WORKER_URL="https://vegvisr-container.torarnehave.workers.dev"

echo "📹 Testing YouTube URL: $VIDEO_URL"
echo "🔧 Instance ID: $INSTANCE_ID"
echo ""

echo "🚀 Attempting to extract audio from YouTube..."

# Test the extraction
RESPONSE=$(curl -s -X POST "$WORKER_URL/ffmpeg/$INSTANCE_ID" \
  -H "Content-Type: application/json" \
  -d "{\"video_url\": \"$VIDEO_URL\"}" \
  --max-time 120)

echo "📄 Response:"
echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"

# Check if extraction was successful
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo ""
    echo "✅ YouTube audio extraction successful!"
    
    # Try to extract download info
    DOWNLOAD_URL=$(echo "$RESPONSE" | jq -r '.download_url // empty' 2>/dev/null)
    FILE_NAME=$(echo "$RESPONSE" | jq -r '.file_name // empty' 2>/dev/null)
    VIDEO_TITLE=$(echo "$RESPONSE" | jq -r '.video_title // empty' 2>/dev/null)
    
    if [ -n "$DOWNLOAD_URL" ]; then
        echo "🎵 Video Title: $VIDEO_TITLE"
        echo "📁 File: $FILE_NAME"
        echo "🔗 Download URL: $WORKER_URL$DOWNLOAD_URL"
        
        echo ""
        echo "💾 Testing download..."
        curl -s "$WORKER_URL$DOWNLOAD_URL" -o "youtube_test_audio.mp3" -w "Downloaded: %{size_download} bytes\n"
        
        if [ -f "youtube_test_audio.mp3" ] && [ -s "youtube_test_audio.mp3" ]; then
            echo "✅ Download successful!"
            echo "📊 File info:"
            file "youtube_test_audio.mp3" 2>/dev/null || echo "Audio file created"
            ls -lh "youtube_test_audio.mp3" 2>/dev/null
        else
            echo "❌ Download failed or file is empty"
        fi
    else
        echo "⚠️  No download URL in response"
    fi
    
else
    echo ""
    echo "❌ YouTube audio extraction failed"
    ERROR=$(echo "$RESPONSE" | jq -r '.error // "Unknown error"' 2>/dev/null)
    echo "💥 Error: $ERROR"
fi

echo ""
echo "🏁 Test completed"