import { serve } from "bun";

const server = serve({
  port: 3000,
  fetch(req) {
    const url = new URL(req.url);
    let filePath = url.pathname;

    if (filePath === "/") {
      filePath = "/index.html";
    }

    const file = Bun.file("." + filePath);
    return new Response(file);
  },
});

console.log(`Server running at http://localhost:${server.port}`);
