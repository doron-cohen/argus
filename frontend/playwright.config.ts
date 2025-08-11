/// <reference types="node" />
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/e2e",
  testMatch: /.*\.(test|spec)\.(js|ts|mjs)/,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
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
    baseURL: "http://localhost:8080", // Point to real application
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
  webServer: process.env.CI
    ? undefined
    : {
        command: "docker-compose up --build -d && sleep 30", // Start full stack
        url: "http://localhost:8080",
        reuseExistingServer: true,
        timeout: 120 * 1000,
      },
  timeout: 60000,
  expect: {
    timeout: 3000,
  },
  outputDir: "test-results/",
});
