import { test, expect, type Page } from "@playwright/test";
import type { Component } from "./types";
import { spawn, execFile } from "child_process";
import { promisify } from "util";

const execFileAsync = promisify(execFile);

async function runSeedScript(args: string[] = []): Promise<void> {
  return new Promise((resolve, reject) => {
    const child = spawn("bun", ["../../scripts/seed-reports.js", ...args], {
      cwd: __dirname,
      stdio: "pipe",
      env: { ...process.env, ARGUS_BASE_URL: "http://localhost:8080" },
    });

    let stdout = "";
    let stderr = "";

    child.stdout?.on("data", (data) => {
      stdout += data.toString();
    });

    child.stderr?.on("data", (data) => {
      stderr += data.toString();
    });

    child.on("close", (code) => {
      if (code === 0) {
        console.log("Seed script output:", stdout);
        resolve();
      } else {
        console.error("Seed script failed:", stderr);
        reject(new Error(`Seed script failed with code ${code}: ${stderr}`));
      }
    });

    child.on("error", (error) => {
      reject(error);
    });
  });
}

async function waitForSync(page: Page): Promise<void> {
  await page.waitForFunction(
    async () => {
      try {
        const response = await fetch(
          "http://localhost:8080/api/sync/v1/sources/0/status"
        );
        const data = await response.json();
        return (
          data.status === "completed" || data.lastSync?.status === "success"
        );
      } catch {
        return false;
      }
    },
    { timeout: 30000 }
  );
}

async function getComponentsFromAPI(): Promise<Component[]> {
  const response = await fetch(
    "http://localhost:8080/api/catalog/v1/components"
  );
  if (!response.ok) {
    throw new Error(`Failed to fetch components: ${response.status}`);
  }
  return response.json();
}

async function getComponentReports(componentId: string): Promise<any> {
  const response = await fetch(
    `http://localhost:8080/api/catalog/v1/components/${componentId}/reports?latest_per_check=true`
  );
  if (!response.ok) {
    if (response.status === 404) {
      return { reports: [] };
    }
    throw new Error(`Failed to fetch reports: ${response.status}`);
  }
  return response.json();
}

test.describe("Component Reports", () => {
  test.beforeEach(async ({ page }: { page: Page }) => {
    // Wait for sync to complete before each test
    await waitForSync(page);
  });

  test.describe("Empty State", () => {
    test("should display no reports message when component has no reports", async ({
      page,
    }: {
      page: Page;
    }) => {
      // Navigate to a component that should have no reports initially
      await page.goto("/components/user-service");

      // Wait for component details to load
      await expect(page.getByTestId("component-details")).toBeVisible();
      await expect(page.getByTestId("component-name")).toContainText(
        "User Service"
      );

      // Check that reports section exists
      await expect(page.getByTestId("reports-label")).toContainText(
        "Latest Quality Checks"
      );

      // Verify no reports message is displayed
      await expect(page.getByTestId("no-reports")).toBeVisible();
      await expect(page.getByTestId("no-reports")).toContainText(
        "No quality checks available"
      );

      // Ensure reports list is not present
      await expect(page.getByTestId("reports-list")).not.toBeVisible();
    });

    test("should not show loading or error states when no reports exist", async ({
      page,
    }: {
      page: Page;
    }) => {
      await page.goto("/components/user-service");

      await expect(page.getByTestId("component-details")).toBeVisible();

      // These elements should not be present in the empty state
      await expect(page.getByTestId("reports-loading")).not.toBeVisible();
      await expect(page.getByTestId("reports-error")).not.toBeVisible();
    });
  });

  test.describe("Populated State", () => {
    test.beforeAll(async () => {
      // Seed reports for testing populated state
      console.log("🌱 Seeding reports for populated state tests...");
      await runSeedScript([
        "--only",
        "auth-service",
        "--all-statuses",
        "--reports-per-component",
        "5",
      ]);
    });

    test("should display reports when component has quality checks", async ({
      page,
    }: {
      page: Page;
    }) => {
      await page.goto("/components/auth-service");

      // Wait for component details to load
      await expect(page.getByTestId("component-details")).toBeVisible();
      await expect(page.getByTestId("component-name")).toContainText(
        "Authentication Service"
      );

      // Check that reports section exists and has content
      await expect(page.getByTestId("reports-label")).toContainText(
        "Latest Quality Checks"
      );

      // Wait for reports to load and verify reports list is visible
      await expect(page.getByTestId("reports-list")).toBeVisible();

      // Check that individual report items are present
      const reportItems = page.getByTestId("report-item");
      await expect(reportItems).toHaveCount(5); // We seeded 5 reports

      // Verify each report has the required elements
      for (let i = 0; i < 5; i++) {
        const reportItem = reportItems.nth(i);
        await expect(reportItem.getByTestId("check-name")).toBeVisible();
        await expect(reportItem.getByTestId("check-status")).toBeVisible();
        await expect(reportItem.getByTestId("check-timestamp")).toBeVisible();
      }

      // Ensure no empty state is shown
      await expect(page.getByTestId("no-reports")).not.toBeVisible();
    });

    test("should display correct status colors and icons", async ({
      page,
    }: {
      page: Page;
    }) => {
      await page.goto("/components/auth-service");

      await expect(page.getByTestId("component-details")).toBeVisible();
      await expect(page.getByTestId("reports-list")).toBeVisible();

      const statusElements = page.getByTestId("check-status");
      const count = await statusElements.count();

      // Verify we have some status elements
      expect(count).toBeGreaterThan(0);

      for (let i = 0; i < count; i++) {
        const statusElement = statusElements.nth(i);
        const statusText = await statusElement.textContent();
        const className = await statusElement.getAttribute("class");

        // Verify that each status has appropriate color coding
        if (statusText?.includes("pass")) {
          expect(className).toContain("bg-green-100");
        } else if (
          statusText?.includes("fail") ||
          statusText?.includes("error") ||
          statusText?.includes("unknown")
        ) {
          expect(className).toContain("bg-red-100");
        } else if (
          statusText?.includes("disabled") ||
          statusText?.includes("skipped")
        ) {
          expect(className).toContain("bg-yellow-100");
        } else if (statusText?.includes("completed")) {
          expect(className).toContain("bg-blue-100");
        }

        // Verify icon is present
        const svg = statusElement.locator("svg").first();
        await expect(svg).toBeVisible();
      }
    });

    test("should display formatted timestamps", async ({
      page,
    }: {
      page: Page;
    }) => {
      await page.goto("/components/auth-service");

      await expect(page.getByTestId("reports-list")).toBeVisible();

      const timestamps = page.getByTestId("check-timestamp");
      const count = await timestamps.count();

      expect(count).toBeGreaterThan(0);

      for (let i = 0; i < count; i++) {
        const timestamp = timestamps.nth(i);
        const timestampText = await timestamp.textContent();

        // Verify timestamp format (should contain date and time)
        expect(timestampText).toBeTruthy();
        expect(timestampText).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}/); // Date format
        expect(timestampText).toMatch(/\d{1,2}:\d{2}/); // Time format
      }
    });

    test("should show different check types", async ({
      page,
    }: {
      page: Page;
    }) => {
      await page.goto("/components/auth-service");

      await expect(page.getByTestId("reports-list")).toBeVisible();

      const checkNames = page.getByTestId("check-name");
      const count = await checkNames.count();

      expect(count).toBeGreaterThan(0);

      const checkNameTexts = [];
      for (let i = 0; i < count; i++) {
        const checkName = checkNames.nth(i);
        const nameText = await checkName.textContent();
        checkNameTexts.push(nameText);
      }

      // Verify we have different check types (should be unique)
      const uniqueCheckNames = new Set(checkNameTexts);
      expect(uniqueCheckNames.size).toBeGreaterThan(1);

      // Verify some expected check types exist
      const expectedChecks = [
        "unit-tests",
        "security-scan",
        "code-quality",
        "build",
        "integration-tests",
      ];
      const foundChecks = checkNameTexts.filter((name) =>
        expectedChecks.includes(name || "")
      );
      expect(foundChecks.length).toBeGreaterThan(0);
    });
  });

  test.describe("API Integration", () => {
    test("should match API data with frontend display", async ({
      page,
    }: {
      page: Page;
    }) => {
      // Get reports from API
      const apiReports = await getComponentReports("auth-service");

      if (apiReports.reports.length === 0) {
        // If no reports, test empty state
        await page.goto("/components/auth-service");
        await expect(page.getByTestId("no-reports")).toBeVisible();
        return;
      }

      // Navigate to component and verify reports match
      await page.goto("/components/auth-service");
      await expect(page.getByTestId("reports-list")).toBeVisible();

      // Count should match
      const reportItems = page.getByTestId("report-item");
      await expect(reportItems).toHaveCount(apiReports.reports.length);

      // Verify each report matches API data
      for (let i = 0; i < apiReports.reports.length; i++) {
        const apiReport = apiReports.reports[i];
        const reportItem = reportItems.nth(i);

        // Check that the check name matches
        const checkNameElement = reportItem.getByTestId("check-name");
        await expect(checkNameElement).toContainText(apiReport.check_slug);

        // Check that the status matches
        const statusElement = reportItem.getByTestId("check-status");
        await expect(statusElement).toContainText(apiReport.status);
      }
    });
  });

  test.describe("Mixed Scenarios", () => {
    test.beforeAll(async () => {
      // Seed reports for some components but not others
      console.log("🌱 Seeding reports for mixed scenario tests...");
      await runSeedScript([
        "--exclude",
        "user-service", // Keep this one empty
        "--reports-per-component",
        "3",
      ]);
    });

    test("should handle components with and without reports correctly", async ({
      page,
    }: {
      page: Page;
    }) => {
      // First visit a component with reports
      await page.goto("/components/auth-service");
      await expect(page.getByTestId("component-details")).toBeVisible();

      // Should have reports
      await expect(page.getByTestId("reports-list")).toBeVisible();
      await expect(page.getByTestId("no-reports")).not.toBeVisible();

      // Navigate to component without reports
      await page.goto("/components/user-service");
      await expect(page.getByTestId("component-details")).toBeVisible();

      // Should show empty state
      await expect(page.getByTestId("no-reports")).toBeVisible();
      await expect(page.getByTestId("reports-list")).not.toBeVisible();

      // Navigate back to component with reports
      await page.goto("/components/auth-service");
      await expect(page.getByTestId("component-details")).toBeVisible();

      // Should show reports again
      await expect(page.getByTestId("reports-list")).toBeVisible();
      await expect(page.getByTestId("no-reports")).not.toBeVisible();
    });
  });

  test.describe("Error Handling", () => {
    test("should handle non-existent component gracefully", async ({
      page,
    }: {
      page: Page;
    }) => {
      await page.goto("/components/non-existent-component");

      // Should show component error, not reports error
      await expect(page.getByTestId("component-details-error")).toBeVisible();
      await expect(page.getByTestId("error-title")).toContainText(
        "Error loading component"
      );
    });
  });
});
