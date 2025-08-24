// API Configuration
// Handles setting up the API host from environment variables

// Export something to make this a module
export {};

declare global {
  interface Window {
    __ARGUS_API_HOST?: string;
  }
  const __ARGUS_API_HOST__: string | undefined;
}

// Set API host from build-time constant
function initializeApiHost(): void {
  // Use build-time constant defined by Bun
  const envApiHost =
    typeof __ARGUS_API_HOST__ !== "undefined"
      ? __ARGUS_API_HOST__
      : "http://localhost:8080";

  // If envApiHost is empty string, use relative URLs (for production build served by backend)
  window.__ARGUS_API_HOST = envApiHost === "" ? "" : envApiHost;
}

// Initialize on module load
initializeApiHost();
