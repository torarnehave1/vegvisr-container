import { Container, getContainer, getRandom } from "@cloudflare/containers";
import { Hono } from "hono";

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
