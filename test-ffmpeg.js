/**
 * Test script for the FFmpeg endpoint
 * This demonstrates how to use the new audio extraction endpoint
 */

const WORKER_URL = "https://your-worker.your-subdomain.workers.dev"; // Replace with your actual worker URL

async function testFFmpegEndpoint() {
    try {
        // Example video URL (you can replace this with any valid video URL)
        const testVideoURL = "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4";
        
        const response = await fetch(`${WORKER_URL}/ffmpeg/test-instance`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                video_url: testVideoURL
            })
        });

        const result = await response.json();
        
        if (result.success) {
            console.log("✅ Audio extraction successful!");
            console.log("Message:", result.message);
            console.log("Audio URL:", result.audio_url);
            
            // If you want to download the audio file
            if (result.audio_url) {
                const audioResponse = await fetch(`${WORKER_URL}/container/test-instance${result.audio_url}`);
                if (audioResponse.ok) {
                    console.log("🎵 Audio file is ready for download");
                }
            }
        } else {
            console.error("❌ Audio extraction failed:", result.error);
        }
        
    } catch (error) {
        console.error("❌ Test failed:", error);
    }
}

// Example usage with curl command
console.log("You can also test with curl:");
console.log(`
curl -X POST ${WORKER_URL}/ffmpeg/test-instance \\
  -H "Content-Type: application/json" \\
  -d '{"video_url": "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4"}'
`);

// Run the test
testFFmpegEndpoint();