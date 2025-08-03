import { serve } from "bun";
import { normalize, relative } from "path";

// Validate and sanitize file paths to prevent directory traversal
function sanitizePath(pathname) {
  // Normalize the path to handle '..' and '.' segments
  const normalizedPath = normalize(pathname);

  // Check if the normalized path tries to escape the current directory
  const relativePath = relative(".", normalizedPath);

  // If the relative path starts with '..' or is absolute, it's trying to escape
  if (relativePath.startsWith("..") || normalizedPath.startsWith("/")) {
    return null; // Invalid path
  }

  return normalizedPath;
}

const server = serve({
  port: 3000,
  fetch(req) {
    const url = new URL(req.url);
    let filePath = url.pathname;

    if (filePath === "/") {
      filePath = "/index.html";
    }

    // Sanitize the file path to prevent directory traversal
    const sanitizedPath = sanitizePath(filePath);
    if (sanitizedPath === null) {
      return new Response("Forbidden", { status: 403 });
    }

    const file = Bun.file("." + sanitizedPath);
    return new Response(file);
  },
});

console.log(`Server running at http://localhost:${server.port}`);
