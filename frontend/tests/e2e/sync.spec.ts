import { test, expect, type Page } from "@playwright/test";
import type { Component, SyncSource, SyncStatus } from "./types";

test.describe("Sync Process", () => {
  test("should complete sync and populate components", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Wait for sync to complete
    await page.waitForFunction(
      async () => {
        try {
          const response = await fetch(
            "http://localhost:8080/api/sync/v1/sources/0/status",
          );
          const data: SyncStatus = await response.json();
          return (
            data.status === "completed" || data.lastSync?.status === "success"
          );
        } catch {
          return false;
        }
      },
      { timeout: 5000 },
    );

    // Verify sync configuration
    const syncResponse = await page.request.get(
      "http://localhost:8080/api/sync/v1/sources",
    );
    expect(syncResponse.ok()).toBeTruthy();

    const syncData: SyncSource[] = await syncResponse.json();
    expect(syncData).toHaveLength(1);
    expect(syncData[0].type).toBe("filesystem");
    expect(syncData[0].config.path).toContain("testdata");

    // Verify components are available via API
    const apiResponse = await page.request.get(
      "http://localhost:8080/api/catalog/v1/components",
    );
    expect(apiResponse.ok()).toBeTruthy();

    const components: Component[] = await apiResponse.json();
    expect(components).toHaveLength(4);

    // Verify specific components from test data
    const componentIds: string[] = components.map((c: Component) => c.id);
    expect(componentIds).toContain("auth-service");
    expect(componentIds).toContain("api-gateway");
    expect(componentIds).toContain("user-service");
    expect(componentIds).toContain("platform-infrastructure");
  });
});
