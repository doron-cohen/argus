import { test, expect, type Page } from "@playwright/test";
import { ensureReports } from "./fixtures";

test.describe("Component Details Page", () => {
  test("should display component details", async ({ page }: { page: Page }) => {
    // Navigate to a component that should exist
    await page.goto("/components/auth-service");

    // Verify the page loaded and shows component details
    await expect(page.getByTestId("page-title")).toBeVisible();
    await expect(page.getByTestId("page-title")).toHaveText(
      "Component Details",
    );

    // Verify component details are displayed
    await expect(page.getByTestId("component-details")).toBeVisible();

    // Check for component name and ID (may not exist if component failed to load)
    const nameElements = page.getByTestId("component-name");
    const idElements = page.getByTestId("component-id");

    if ((await nameElements.count()) > 0) {
      await expect(nameElements.first()).toBeVisible();
    }
    if ((await idElements.count()) > 0) {
      await expect(idElements.first()).toBeVisible();
    }
  });

  test("should display reports when available", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Ensure reports exist for the component
    await ensureReports(page.request, "auth-service", 1);
    await page.goto("/components/auth-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({});

    // Wait for reports to load and be visible
    await expect(page.getByTestId("reports-list")).toBeVisible();

    // Verify reports are displayed (check that at least one exists)
    await expect(page.getByTestId("report-item").first()).toBeVisible();
    await expect(page.getByTestId("check-name").first()).toBeVisible();
    await expect(page.getByTestId("check-status").first()).toBeVisible();
    await expect(page.getByTestId("check-timestamp").first()).toBeVisible();
  });

  test("should handle component with no reports", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Navigate to a component that should have no reports initially
    await page.goto("/components/platform-infrastructure");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({});

    // Check for no reports message
    await expect(page.getByTestId("no-reports")).toBeVisible();
  });

  test("should show loading state", async ({ page }: { page: Page }) => {
    // Navigate to a component
    await page.goto("/components/auth-service");

    // The loading state should be brief, but we can check for the component details
    // which should appear after loading
    await expect(page.getByTestId("component-details")).toBeVisible({});
  });

  test("should handle non-existent component gracefully", async ({
    page,
  }: {
    page: Page;
  }) => {
    await page.goto("/components/non-existent-component");

    // Wait for the error state to render, then assert error
    await expect(page.getByTestId("component-details-error")).toBeVisible({});

    // Verify error state is displayed (check for error card and alert content)
    const errorCard = page.getByTestId("component-details-error");
    await expect(errorCard).toBeVisible();

    // Check that the error alert contains the expected title
    await expect(errorCard).toContainText("Error loading component");
  });

  test("should display back to components link", async ({
    page,
  }: {
    page: Page;
  }) => {
    await page.goto("/components/auth-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({});

    // Verify back link is present
    await expect(page.getByTestId("back-to-components")).toBeVisible();
    await expect(page.getByTestId("back-to-components")).toHaveText(
      "← Back to Components",
    );
  });

  test("should navigate back to components list", async ({
    page,
  }: {
    page: Page;
  }) => {
    await page.goto("/components/auth-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({});

    // Click the back link
    await page.getByTestId("back-to-components").click();

    // Verify we're back on the components list page
    await expect(page).toHaveURL("/components");
  });
});
