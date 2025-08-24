import { test, expect } from "@playwright/test";

test.describe("Settings Page", () => {
  test("should load settings page", async ({ page }) => {
    // Navigate to settings page
    await page.goto("/settings");
    await page.waitForLoadState("load");

    // Basic page structure should be present
    await expect(page.getByTestId("page-title")).toBeVisible();
    await expect(page.getByTestId("page-title")).toHaveText("Settings");
  });

  test("should display page description", async ({ page }) => {
    await page.goto("/settings");
    await page.waitForLoadState("load");

    await expect(page.getByTestId("page-description")).toBeVisible();
    await expect(page.getByTestId("page-description")).toHaveText(
      "Sync source configuration and status information",
    );
  });

  test("should show content after loading", async ({ page }) => {
    await page.goto("/settings");
    await page.waitForLoadState("load");

    // Wait for the app to be ready
    await expect(page.getByTestId("page-title")).toBeVisible();

    // Wait for API calls to complete and content to load
    await expect(page.getByTestId("sync-source-unknown")).toBeVisible();

    // Check that filesystem content is visible
    await expect(page.getByText("Filesystem")).toBeVisible();
  });
});
