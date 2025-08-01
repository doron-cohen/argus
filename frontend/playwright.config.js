import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI
    ? [
        ["json", { outputFile: "test-results/results.json" }],
        ["junit", { outputFile: "test-results/results.xml" }],
      ]
    : [
        ["html"],
        ["json", { outputFile: "test-results/results.json" }],
        ["junit", { outputFile: "test-results/results.xml" }],
      ],
  use: {
    baseURL: "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    // Global timeout for actions
    actionTimeout: 10000,
    // Global timeout for navigation
    navigationTimeout: 30000,
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "firefox",
      use: { ...devices["Desktop Firefox"] },
    },
    {
      name: "webkit",
      use: { ...devices["Desktop Safari"] },
    },
  ],
  ...(process.env.CI
    ? {}
    : {
        webServer: {
          command: "bun server.js",
          url: "http://localhost:3000",
          reuseExistingServer: !process.env.CI,
          timeout: 120 * 1000,
        },
      }),
  // Global test timeout
  timeout: 60000,
  // Expect timeout for assertions
  expect: {
    timeout: 10000,
  },
  // Output directory for test artifacts
  outputDir: "test-results/",
});
