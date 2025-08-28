import { defineConfig } from "vite";

export default defineConfig({
  root: ".",
  build: {
    outDir: "dist",
    rollupOptions: {
      input: {
        main: "./index.html",
      },
    },
  },
  server: {
    port: 3000,
    proxy: {
      "/api": {
        target: process.env.VITE_API_HOST || "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  define: {
    __ARGUS_API_HOST__: JSON.stringify(
      process.env.VITE_API_HOST || "http://localhost:8080"
    ),
  },
});
