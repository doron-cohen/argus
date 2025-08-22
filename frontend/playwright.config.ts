/// <reference types="node" />
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/e2e",
  testMatch: /.*\.(test|spec)\.(js|ts|mjs)/,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI
    ? [
        ["json", { outputFile: "test-results/results.json" }],
        ["junit", { outputFile: "test-results/results.xml" }],
        ["html", { outputFolder: "playwright-report", open: "never" }],
      ]
    : [
        ["list"],
        ["json", { outputFile: "test-results/results.json" }],
        ["junit", { outputFile: "test-results/results.xml" }],
        ["html", { outputFolder: "playwright-report", open: "never" }],
      ],
  use: {
    baseURL: process.env.BASE_URL || "http://localhost:8080", // Allows dev server at :3000
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    actionTimeout: 10000,
    navigationTimeout: 30000,
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  // No webServer config - we handle Docker startup manually in Makefile targets
  timeout: 60000,
  expect: {
    timeout: 7000,
  },
  outputDir: "test-results/",
});
