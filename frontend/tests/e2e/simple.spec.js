import { test, expect } from "@playwright/test";

test.describe("Simple Page Test", () => {
  test("should load the test HTML file directly", async ({ page }) => {
    // Given: I navigate to the test HTML file
    await page.goto("file://" + process.cwd() + "/test.html");

    // When: The page loads
    await page.waitForLoadState("networkidle");

    // Then: I should see the page title
    await expect(page.locator('[data-testid="page-title"]')).toContainText(
      "Component Catalog Test"
    );

    // And: I should see the loading spinner initially
    await expect(page.locator('[data-testid="loading-spinner"]')).toBeVisible();

    // And: After loading, I should see the components
    await page.waitForTimeout(1000); // Wait for Alpine.js to load
    await expect(
      page.locator('[data-testid="components-container"]')
    ).toBeVisible();

    // And: I should see the dummy components
    const componentRows = page.locator('[data-testid="component-row"]');
    await expect(componentRows).toHaveCount(3);
  });
});
