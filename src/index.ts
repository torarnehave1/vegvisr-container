import { Container, getContainer, getRandom } from "@cloudflare/containers";
import { Hono } from "hono";
import { cors } from "hono/cors";

export class MyContainer extends Container<Env> {
	// Port the container listens on (default: 8080)
	defaultPort = 8080;
	// Time before container sleeps due to inactivity (default: 30s)
	sleepAfter = "2m";
	// Environment variables passed to the container
	envVars = {
		MESSAGE: "Container with 200MB chunked download support - v2!",
	};

	// Optional lifecycle hooks
	override onStart() {
		console.log("Container successfully started");
	}

	override onStop() {
		console.log("Container successfully shut down");
	}

	override onError(error: unknown) {
		console.log("Container error:", error);
	}
}

// Create Hono app with proper typing for Cloudflare Workers
const app = new Hono<{
	Bindings: Env;
}>();

// Add CORS middleware to allow requests from any origin
app.use('/*', cors({
	origin: '*',
	allowMethods: ['GET', 'POST', 'OPTIONS'],
	allowHeaders: ['Content-Type'],
	maxAge: 86400,
}));

// Home route with available endpoints
app.get("/", (c) => {
	return c.text(
		"Available endpoints:\n" +
			"GET /container/<ID> - Start a container for each ID with a 2m timeout\n" +
			"GET /lb - Load balance requests over multiple containers\n" +
			"GET /error - Start a container that errors (demonstrates error handling)\n" +
			"GET /singleton - Get a single specific container instance\n" +
			"POST /ffmpeg/<ID> - Extract MP3 audio from video using FFmpeg\n" +
			"  • Supports files up to 200MB with chunked downloading\n" +
			"  • Progress tracking and detailed file size information\n" +
			"  • Automatic R2 storage for extracted audio\n" +
			"POST /upload/<ID> - Upload video files directly (auto-handles small & large files up to 100MB)\n" +
			"  • Smart file size detection (≤15MB: direct processing, >15MB: R2-based)\n" +
			"  • Perfect for QuickTime screen recordings and presentations\n" +
			"GET /download/<filename> - Download extracted audio files from R2 storage",
	);
});

// Route requests to a specific container using the container ID
app.get("/container/:id", async (c) => {
	const id = c.req.param("id");
	const containerId = c.env.MY_CONTAINER.idFromName(`/container/${id}`);
	const container = c.env.MY_CONTAINER.get(containerId);
	return await container.fetch(c.req.raw);
});

// Demonstrate error handling - this route forces a panic in the container
app.get("/error", async (c) => {
	const container = getContainer(c.env.MY_CONTAINER, "error-test");
	return await container.fetch(c.req.raw);
});

// Load balance requests across multiple containers
app.get("/lb", async (c) => {
	const container = await getRandom(c.env.MY_CONTAINER, 3);
	return await container.fetch(c.req.raw);
});

// Get a single container instance (singleton pattern)
app.get("/singleton", async (c) => {
	const container = getContainer(c.env.MY_CONTAINER);
	return await container.fetch(c.req.raw);
});

// FFmpeg endpoint - extract MP3 from video with R2 storage
app.post("/ffmpeg/:id", async (c) => {
	const id = c.req.param("id");
	const containerId = c.env.MY_CONTAINER.idFromName(`/ffmpeg/${id}`);
	const container = c.env.MY_CONTAINER.get(containerId);
	
	// Create a new request for the container's FFmpeg endpoint
	const body = await c.req.text();
	const requestData = JSON.parse(body);
	
	// Add R2 storage configuration to the request
	const enhancedRequest = {
		...requestData,
		use_r2_storage: true,
		instance_id: id
	};
	
	const newRequest = new Request("http://localhost:8080/ffmpeg/extract-audio", {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify(enhancedRequest),
	});
	
	const containerResponse = await container.fetch(newRequest);
	const result = await containerResponse.json() as {
		success: boolean;
		message?: string;
		audio_data?: string;
		error?: string;
	};
	
	// If audio was extracted successfully, upload to R2
	if (result.success && result.audio_data) {
		try {
			// Create a unique filename
			const timestamp = Date.now();
			const fileName = `audio_${id}_${timestamp}.mp3`;
			
			// Upload to R2
			const audioBuffer = Uint8Array.from(atob(result.audio_data), c => c.charCodeAt(0));
			await c.env.AUDIO_STORAGE.put(fileName, audioBuffer, {
				httpMetadata: {
					contentType: "audio/mpeg"
				}
			});
			
			// Return download URL
			return c.json({
				success: true,
				message: "Audio extracted and stored successfully",
				download_url: `/download/${fileName}`,
				file_name: fileName,
				r2_key: fileName
			});
		} catch (error: any) {
			return c.json({
				success: false,
				error: `R2 storage failed: ${error.message}`
			});
		}
	}
	
	return c.json(result);
});

// Upload endpoint - unified smart upload that handles all file sizes
app.post("/upload/:id", async (c) => {
	const id = c.req.param("id");
	const containerId = c.env.MY_CONTAINER.idFromName(`/upload/${id}`);
	const container = c.env.MY_CONTAINER.get(containerId);
	
	try {
		// Parse form data
		const formData = await c.req.formData();
		const videoFile = (formData.get('video') || formData.get('file')) as File;
		const outputFormat = formData.get('output_format') as string || 'mp3';
		
		if (!videoFile) {
			return c.json({
				success: false,
				error: "No video file provided (use 'video' or 'file' field name)"
			});
		}
		
		const fileSizeMB = videoFile.size / (1024 * 1024);
		console.log(`Processing file: ${videoFile.name}, size: ${fileSizeMB.toFixed(1)}MB`);
		
		// Check maximum file size
		if (videoFile.size > 100 * 1024 * 1024) {
			return c.json({
				success: false,
				error: `File too large (${fileSizeMB.toFixed(1)}MB). Maximum supported: 100MB`
			});
		}
		
		// Smart routing based on file size
		// For large files (> 15MB), upload to R2 first to avoid memory issues
		if (videoFile.size > 15 * 1024 * 1024) {
			console.log(`Large file detected (${fileSizeMB.toFixed(1)}MB), uploading to R2 first`);
			
			// Generate unique R2 key
			const timestamp = Date.now();
			const r2Key = `temp-uploads/${id}/${timestamp}_${videoFile.name}`;
			
			// Upload to R2 first
			await c.env.AUDIO_STORAGE.put(r2Key, videoFile.stream(), {
				httpMetadata: {
					contentType: videoFile.type || 'video/quicktime'
				}
			});
			
			console.log(`File uploaded to R2: ${r2Key}`);
			
			// Get the file back from R2
			const object = await c.env.AUDIO_STORAGE.get(r2Key);
			
			if (!object) {
				return c.json({
					success: false,
					error: "File upload to R2 failed"
				});
			}
			
			// Download file from R2 in chunks
			const videoData = await object.arrayBuffer();
			
			// Convert to base64 in chunks
			const uint8Array = new Uint8Array(videoData);
			let binaryString = '';
			const chunkSize = 8192;
			for (let i = 0; i < uint8Array.length; i += chunkSize) {
				const chunk = uint8Array.subarray(i, Math.min(i + chunkSize, uint8Array.length));
				binaryString += String.fromCharCode.apply(null, Array.from(chunk));
			}
			const videoBase64 = btoa(binaryString);
			
			// Send to container for processing
			const containerRequest = {
				video_data: videoBase64,
				filename: videoFile.name,
				file_size: videoFile.size,
				output_format: outputFormat,
				instance_id: id
			};
			
			const newRequest = new Request("http://localhost:8080/ffmpeg/upload-base64", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(containerRequest),
			});
			
			const containerResponse = await container.fetch(newRequest);
			
			if (!containerResponse.ok) {
				const errorText = await containerResponse.text();
				await c.env.AUDIO_STORAGE.delete(r2Key);
				return c.json({
					success: false,
					error: `Container returned ${containerResponse.status}: ${errorText}`
				});
			}
			
			const result = await containerResponse.json() as {
				success: boolean;
				message?: string;
				download_url?: string;
				file_name?: string;
				r2_key?: string;
				audio_url?: string;
				audio_data?: string;
				error?: string;
			};
			
			// If audio was extracted, upload to R2
			if (result.success && result.audio_data && result.file_name) {
				try {
					const audioBuffer = Uint8Array.from(atob(result.audio_data), c => c.charCodeAt(0));
					await c.env.AUDIO_STORAGE.put(result.file_name, audioBuffer, {
						httpMetadata: {
							contentType: "audio/mpeg"
						}
					});
					
					// Clean up the temp video file from R2
					await c.env.AUDIO_STORAGE.delete(r2Key);
					
					return c.json({
						success: true,
						message: result.message,
						download_url: result.download_url,
						file_name: result.file_name,
						r2_key: result.r2_key,
						audio_url: result.audio_url
					});
				} catch (error: any) {
					await c.env.AUDIO_STORAGE.delete(r2Key);
					return c.json({
						success: false,
						error: `R2 storage failed: ${error.message}`
					});
				}
			}
			
			await c.env.AUDIO_STORAGE.delete(r2Key);
			return c.json(result);
			
		} else {
			// Small files (≤ 15MB) - direct base64 encoding approach
			console.log(`Small file (${fileSizeMB.toFixed(1)}MB), using direct processing`);
			
			// Convert file to base64 using chunked approach
			const videoBuffer = await videoFile.arrayBuffer();
			const uint8Array = new Uint8Array(videoBuffer);
			
			// Convert to binary string in chunks to avoid call stack size exceeded
			let binaryString = '';
			const chunkSize = 8192;
			for (let i = 0; i < uint8Array.length; i += chunkSize) {
				const chunk = uint8Array.subarray(i, Math.min(i + chunkSize, uint8Array.length));
				binaryString += String.fromCharCode.apply(null, Array.from(chunk));
			}
			
			const videoBase64 = btoa(binaryString);
			
			// Send to container as JSON
			const containerRequest = {
				video_data: videoBase64,
				filename: videoFile.name,
				file_size: videoFile.size,
				output_format: outputFormat,
				instance_id: id
			};
			
			const newRequest = new Request("http://localhost:8080/ffmpeg/upload-base64", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(containerRequest),
			});
			
			const containerResponse = await container.fetch(newRequest);
			
			if (!containerResponse.ok) {
				const errorText = await containerResponse.text();
				return c.json({
					success: false,
					error: `Container returned ${containerResponse.status}: ${errorText}`
				});
			}
			
			const result = await containerResponse.json() as {
				success: boolean;
				message?: string;
				download_url?: string;
				file_name?: string;
				r2_key?: string;
				audio_url?: string;
				audio_data?: string;
				error?: string;
			};
			
			// If audio was extracted successfully, upload to R2
			if (result.success && result.audio_data && result.file_name) {
				try {
					const audioBuffer = Uint8Array.from(atob(result.audio_data), c => c.charCodeAt(0));
					await c.env.AUDIO_STORAGE.put(result.file_name, audioBuffer, {
						httpMetadata: {
							contentType: "audio/mpeg"
						}
					});
					
					return c.json({
						success: true,
						message: result.message,
						download_url: result.download_url,
						file_name: result.file_name,
						r2_key: result.r2_key,
						audio_url: result.audio_url
					});
				} catch (error: any) {
					return c.json({
						success: false,
						error: `R2 storage failed: ${error.message}`
					});
				}
			}
			
			return c.json(result);
		}
		
	} catch (error: any) {
		return c.json({
			success: false,
			error: `Upload processing failed: ${error.message}`
		});
	}
});

// Download endpoint for R2-stored files
app.get("/download/:filename", async (c) => {
	const filename = c.req.param("filename");
	
	try {
		const object = await c.env.AUDIO_STORAGE.get(filename);
		
		if (object === null) {
			return c.json({ error: "File not found" }, 404);
		}
		
		const headers = new Headers();
		headers.set("Content-Type", "audio/mpeg");
		headers.set("Content-Disposition", `attachment; filename="${filename}"`);
		
		return new Response(object.body, { headers });
	} catch (error: any) {
		return c.json({ error: `Download failed: ${error.message}` }, 500);
	}
});

// YouTube info endpoint
app.post("/youtube/:id/info", async (c) => {
	const id = c.req.param("id");
	const containerId = c.env.MY_CONTAINER.idFromName(`/youtube/${id}`);
	const container = c.env.MY_CONTAINER.get(containerId);
	
	// Create a new request for the container's YouTube info endpoint
	const body = await c.req.text();
	const newRequest = new Request("http://localhost:8080/youtube/info", {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: body,
	});
	
	return await container.fetch(newRequest);
});

export default app;
