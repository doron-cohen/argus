import { test, expect, type Page } from "@playwright/test";
import type { Component, SyncSource, SyncStatus } from "./types";

test.describe("Component Catalog - Real Application Flow", () => {
  test.beforeEach(async ({ page }: { page: Page }) => {
    // Wait for sync to complete before each test
    await page.waitForFunction(
      async () => {
        try {
          const response = await fetch(
            "http://localhost:8080/api/sync/v1/sources/0/status"
          );
          const data: SyncStatus = await response.json();
          return (
            data.status === "completed" || data.lastSync?.status === "success"
          );
        } catch {
          return false;
        }
      },
      { timeout: 30000 }
    );
  });

  test("should display all test components from sync process", async ({
    page,
  }: {
    page: Page;
  }) => {
    await page.goto("/");

    // Wait for components to load and verify count
    await expect(page.getByTestId("component-row")).toHaveCount(4);
    await expect(page.getByTestId("components-header")).toContainText(
      "Components (4)"
    );

    // Verify specific components from test data
    const components = page.getByTestId("component-row");

    // Check Authentication Service
    await expect(components.nth(0).getByTestId("component-name")).toContainText(
      "Authentication Service"
    );
    await expect(components.nth(0).getByTestId("component-team")).toContainText(
      "Security Team"
    );
    await expect(
      components.nth(0).getByTestId("component-maintainers")
    ).toContainText("alice@company.com");

    // Check Platform Infrastructure (index 1)
    await expect(components.nth(1).getByTestId("component-name")).toContainText(
      "Platform Infrastructure"
    );
    await expect(components.nth(1).getByTestId("component-team")).toContainText(
      "Infrastructure Team"
    );

    // Check API Gateway (index 2)
    await expect(components.nth(2).getByTestId("component-name")).toContainText(
      "API Gateway"
    );
    await expect(components.nth(2).getByTestId("component-team")).toContainText(
      "Platform Team"
    );
  });

  test("should verify sync API configuration and status", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Test sync sources endpoint
    const sourcesResponse = await page.request.get(
      "http://localhost:8080/api/sync/v1/sources"
    );
    expect(sourcesResponse.ok()).toBeTruthy();

    const sourcesData: SyncSource[] = await sourcesResponse.json();
    expect(sourcesData).toHaveLength(1);
    expect(sourcesData[0].type).toBe("filesystem");
    expect(sourcesData[0].config.path).toContain("testdata");
    expect(sourcesData[0].interval).toBe("30s");

    // Test sync status endpoint
    const statusResponse = await page.request.get(
      "http://localhost:8080/api/sync/v1/sources/0/status"
    );
    expect(statusResponse.ok()).toBeTruthy();
  });

  test("should verify real API responses match frontend display", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Get components via API
    const apiResponse = await page.request.get(
      "http://localhost:8080/api/catalog/v1/components"
    );
    expect(apiResponse.ok()).toBeTruthy();

    const components: Component[] = await apiResponse.json();
    expect(components).toHaveLength(4);

    // Verify component structure
    const authService = components.find(
      (c: Component) => c.id === "auth-service"
    );
    expect(authService?.name).toBe("Authentication Service");
    expect(authService?.owners?.team).toBe("Security Team");
    expect(authService?.owners?.maintainers).toContain("alice@company.com");

    // Verify frontend displays same data
    await page.goto("/");
    await expect(page.getByTestId("component-row")).toHaveCount(4);
    await expect(page.getByTestId("component-name").first()).toContainText(
      "Authentication Service"
    );
  });
});
