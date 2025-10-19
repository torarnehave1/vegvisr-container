#!/bin/bash

# Setup script for deploying the FFmpeg container

echo "🔍 Checking Docker installation..."

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed or not in PATH"
    echo ""
    echo "Please install Docker first:"
    echo "1. Docker Desktop: https://www.docker.com/products/docker-desktop/"
    echo "2. Or via Homebrew: brew install --cask docker"
    echo "3. Or lightweight alternative: brew install colima docker && colima start"
    exit 1
fi

echo "✅ Docker CLI found"

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    echo "❌ Docker daemon is not running"
    echo ""
    echo "Please start Docker:"
    echo "- If using Docker Desktop, open the Docker app"
    echo "- If using Colima, run: colima start"
    exit 1
fi

echo "✅ Docker daemon is running"

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    npm install
fi

echo "🚀 Deploying to Cloudflare..."
npm run deploy

echo ""
echo "🎉 Deployment complete!"
echo ""
echo "You can now test the FFmpeg endpoint:"
echo ""
echo "curl -X POST https://your-worker.workers.dev/ffmpeg/test-instance \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{\"video_url\": \"https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4\"}'"