#!/bin/bash

# Test script for the deployed FFmpeg endpoint
WORKER_URL="https://vegvisr-container.torarnehave.workers.dev"
CONTAINER_ID="test-instance"

echo "🎬 Testing FFmpeg Audio Extraction"
echo "=================================="
echo "Worker URL: $WORKER_URL"
echo "Container ID: $CONTAINER_ID"
echo ""

# Test with a sample video URL (you can replace this with any valid video URL)
VIDEO_URL="https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4"

echo "📹 Video URL: $VIDEO_URL"
echo ""
echo "🚀 Extracting audio..."

# Make the request
response=$(curl -s -X POST "$WORKER_URL/ffmpeg/$CONTAINER_ID" \
  -H "Content-Type: application/json" \
  -d "{\"video_url\": \"$VIDEO_URL\"}")

echo "📄 Response:"
echo "$response" | jq . 2>/dev/null || echo "$response"

# Check if extraction was successful
if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
    echo ""
    echo "✅ Audio extraction successful!"
    
    # Extract audio URL from response
    audio_url=$(echo "$response" | jq -r '.audio_url')
    if [ "$audio_url" != "null" ] && [ "$audio_url" != "" ]; then
        echo "🎵 Audio available at: $WORKER_URL/container/$CONTAINER_ID$audio_url"
        echo ""
        echo "💡 Download with:"
        echo "curl \"$WORKER_URL/container/$CONTAINER_ID$audio_url\" -o extracted_audio.mp3"
    fi
else
    echo ""
    echo "❌ Audio extraction failed"
fi