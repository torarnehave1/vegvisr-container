#!/bin/bash

# Complete workflow demonstration
echo "🎬 FFmpeg Audio Extraction Workflow Demo"
echo "========================================"

WORKER_URL="https://vegvisr-container.torarnehave.workers.dev"
INSTANCE_ID="workflow-demo-$(date +%s)"
VIDEO_URL="https://www.learningcontainer.com/wp-content/uploads/2020/05/sample-mp4-file.mp4"

echo "📋 Configuration:"
echo "   Worker URL: $WORKER_URL"
echo "   Instance ID: $INSTANCE_ID"
echo "   Video URL: $VIDEO_URL"
echo ""

echo "🚀 Step 1: Extracting audio..."
RESPONSE=$(curl -s -X POST "$WORKER_URL/ffmpeg/$INSTANCE_ID" \
  -H "Content-Type: application/json" \
  -d "{\"video_url\": \"$VIDEO_URL\"}")

echo "📄 Response: $RESPONSE"

# Parse the response
SUCCESS=$(echo "$RESPONSE" | grep -o '"success":[^,]*' | cut -d':' -f2)
if [ "$SUCCESS" = "true" ]; then
    echo "✅ Audio extraction successful!"
    
    # Extract the audio URL
    AUDIO_URL=$(echo "$RESPONSE" | grep -o '"/download/[^"]*"' | tr -d '"')
    echo "🎵 Audio URL: $AUDIO_URL"
    
    # Full download URL
    FULL_URL="$WORKER_URL/container/$INSTANCE_ID$AUDIO_URL"
    echo "🔗 Full download URL: $FULL_URL"
    
    echo ""
    echo "💾 Step 2: Downloading audio file (immediately)..."
    
    # Download immediately before cleanup
    curl "$FULL_URL" -o "extracted_audio_$(date +%s).mp3" -w "Downloaded: %{size_download} bytes\n"
    
    if [ $? -eq 0 ]; then
        echo "✅ Download successful!"
        echo "📁 Audio file saved locally"
        ls -la extracted_audio_*.mp3 | tail -1
    else
        echo "❌ Download failed"
    fi
    
else
    echo "❌ Audio extraction failed"
    ERROR=$(echo "$RESPONSE" | grep -o '"error":"[^"]*"' | cut -d':' -f2- | tr -d '"')
    echo "Error: $ERROR"
fi

echo ""
echo "📝 Summary of File Storage:"
echo "   1. Video downloaded to: /tmp/processing/video_{instanceId}_{timestamp}.tmp (inside container)"
echo "   2. Audio extracted to: /tmp/processing/audio_{instanceId}_{timestamp}.mp3 (inside container)"
echo "   3. Audio served via: /download/{filename} endpoint"
echo "   4. Full URL: $WORKER_URL/container/{instance-id}/download/{filename}"
echo "   5. Cleanup: Files auto-deleted after 5 minutes"