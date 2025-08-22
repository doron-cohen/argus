import { serve } from "bun";

const server = serve({
  port: 3000,
  async fetch(req) {
    const url = new URL(req.url);
    let filePath = url.pathname;

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
