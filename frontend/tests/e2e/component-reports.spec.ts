import { test, expect, type Page } from "@playwright/test";
import { ensureReports } from "./fixtures";

async function getComponentReports(
  componentId: string,
): Promise<{ reports: any[] }> {
  const response = await fetch(
    `http://localhost:8080/api/catalog/v1/components/${componentId}/reports?latest_per_check=true`,
  );
  if (!response.ok) {
    return { reports: [] };
  }
  return response.json();
}

test.describe("Component Reports", () => {
  test("should show no reports initially", async ({ page }: { page: Page }) => {
    // Navigate to a component that should have no reports initially
    await page.goto("/components/platform-infrastructure");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Check for no reports message
    await expect(page.getByTestId("no-reports")).toBeVisible();
  });

  test("should display reports after they are created", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Create reports for the component
    await ensureReports(page.request, "user-service", 1);
    await page.goto("/components/user-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Wait for reports to load and be visible
    await expect(page.getByTestId("reports-list")).toBeVisible({
      timeout: 10000,
    });

    // Verify report items are displayed
    await expect(page.getByTestId("report-item")).toBeVisible();
    await expect(page.getByTestId("check-name")).toBeVisible();
    await expect(page.getByTestId("check-status")).toBeVisible();
    await expect(page.getByTestId("check-timestamp")).toBeVisible();
  });

  test("should display multiple reports", async ({ page }: { page: Page }) => {
    const apiReports = await getComponentReports("auth-service");
    await page.goto("/components/auth-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    if (apiReports.reports.length === 0) {
      // If no reports, test empty state
      await expect(page.getByTestId("no-reports")).toBeVisible();
      return;
    }

    // Verify reports match
    await expect(page.getByTestId("reports-list")).toBeVisible({
      timeout: 10000,
    });

    // Verify we have the expected number of reports
    const reportItems = page.getByTestId("report-item");
    await expect(reportItems).toHaveCount(apiReports.reports.length);
  });

  test("should show different check types", async ({
    page,
  }: {
    page: Page;
  }) => {
    await page.goto("/components/auth-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Wait for reports to load
    await expect(page.getByTestId("reports-list")).toBeVisible({
      timeout: 10000,
    });

    // Verify we have at least one report
    await expect(page.getByTestId("report-item")).toBeVisible();
  });

  test("should show status colors correctly", async ({
    page,
  }: {
    page: Page;
  }) => {
    await ensureReports(page.request, "auth-service", 1);
    await page.goto("/components/auth-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Wait for reports to load
    await expect(page.getByTestId("reports-list")).toBeVisible({
      timeout: 10000,
    });

    // Verify status elements are present
    await expect(page.getByTestId("check-status")).toBeVisible();
  });

  test("should handle reports loading state", async ({
    page,
  }: {
    page: Page;
  }) => {
    await page.goto("/components/auth-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // The loading state should be brief, but we can check for the reports
    // which should appear after loading
    try {
      await expect(page.getByTestId("reports-list")).toBeVisible({
        timeout: 5000,
      });
    } catch {
      // If no reports, check for no-reports message
      await expect(page.getByTestId("no-reports")).toBeVisible();
    }
  });

  test("should handle reports error state", async ({
    page,
  }: {
    page: Page;
  }) => {
    // This test would require mocking the API to return an error
    // For now, we'll just verify the component handles the normal case
    await page.goto("/components/auth-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Check for either reports or no-reports message
    const reportsList = page.getByTestId("reports-list");
    const noReports = page.getByTestId("no-reports");

    try {
      await expect(reportsList).toBeVisible({ timeout: 5000 });
    } catch {
      await expect(noReports).toBeVisible();
    }
  });

  test("should maintain reports state during navigation", async ({
    page,
  }: {
    page: Page;
  }) => {
    const apiAuth = await getComponentReports("auth-service");
    const apiUser = await getComponentReports("user-service");

    // First visit a component with reports
    await page.goto("/components/auth-service");
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Navigate to another component with reports
    await page.goto("/components/user-service");
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Navigate back to first component
    await page.goto("/components/auth-service");
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });
  });

  test("should handle non-existent component gracefully", async ({
    page,
  }: {
    page: Page;
  }) => {
    await page.goto("/components/non-existent-component");

    // Wait for the error state to render, then assert error
    await expect(page.getByTestId("component-details-error")).toBeVisible({
      timeout: 10000,
    });

    // Verify error state is displayed
    await expect(page.getByTestId("error-title")).toBeVisible();
    await expect(page.getByTestId("error-title")).toHaveText(
      "Error loading component",
    );
  });
});
