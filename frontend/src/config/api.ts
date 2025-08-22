// API Configuration
// Handles setting up the API host from environment variables

// Export something to make this a module
export {};

declare global {
  interface Window {
    __ARGUS_API_HOST?: string;
  }
}

// Set API host from environment variable
function initializeApiHost(): void {
  const envApiHost =
    import.meta.env?.VITE_API_HOST || process.env?.VITE_API_HOST;

  if (envApiHost) {
    window.__ARGUS_API_HOST = envApiHost;
  }
}

// Initialize on module load
initializeApiHost();
