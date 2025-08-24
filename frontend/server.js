import { serve } from "bun";

const API_HOST = process.env.VITE_API_HOST || "http://localhost:8080";

const server = serve({
  port: 3000,
  async fetch(req) {
    const url = new URL(req.url);
    let filePath = url.pathname;

    // Proxy API calls to backend
    if (filePath.startsWith("/api/")) {
      const apiUrl = `${API_HOST}${filePath}`;
      const apiResponse = await fetch(apiUrl, {
        method: req.method,
        headers: req.headers,
        body: req.body,
      });
      return apiResponse;
    }

    // Handle client-side routing - serve index.html for app routes
    // Serve index.html for any non-dist path (SPA)
    if (!filePath.startsWith("/dist/")) {
      filePath = "/index.html";
    }

    const file = Bun.file("." + filePath);
    return new Response(file);
  },
});

console.log(`Server running at http://localhost:${server.port}`);
