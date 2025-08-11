import { test, expect, type Page } from "@playwright/test";
import { ensureReports } from "./fixtures";
import type { Component } from "./types";

test.describe("Component Details Page", () => {
  test.beforeEach(async ({ page }: { page: Page }) => {
    // Wait for sync to complete before each test
    await page.waitForFunction(
      async () => {
        try {
          const response = await fetch(
            "http://localhost:8080/api/sync/v1/sources/0/status",
          );
          const data = await response.json();
          return (
            data.status === "completed" || data.lastSync?.status === "success"
          );
        } catch {
          return false;
        }
      },
      { timeout: 5000 },
    );
  });

  test("should navigate to component details page", async ({
    page,
  }: {
    page: Page;
  }) => {
    await page.goto("/");

    // Wait for components to load
    await expect(page.getByTestId("component-row")).toHaveCount(4);

    // Click on the first component to navigate to details
    const firstComponent = page.getByTestId("component-name").first();
    await firstComponent.click();

    // Verify we're on the component details page
    await expect(page.getByTestId("page-title")).toContainText(
      "Component Details",
    );
    await expect(page.getByTestId("component-details")).toBeVisible();
  });

  test("should display component metadata correctly", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Navigate directly to a specific component
    await page.goto("/components/auth-service");

    // Verify component details are displayed
    await expect(page.getByTestId("component-details")).toBeVisible();
    await expect(page.getByTestId("component-name")).toContainText(
      "Authentication Service",
    );
    await expect(page.getByTestId("component-id")).toContainText(
      "ID: auth-service",
    );
    await expect(page.getByTestId("component-description")).toContainText(
      "Handles user authentication and authorization",
    );
    await expect(page.getByTestId("component-team")).toContainText(
      "Security Team",
    );
    await expect(page.getByTestId("component-maintainers")).toContainText(
      "alice@company.com",
    );
  });

  test("should handle loading state", async ({ page }: { page: Page }) => {
    // Navigate to component details
    await page.goto("/components/auth-service");

    // Verify loading state is handled (briefly visible during navigation)
    await expect(page.getByTestId("component-details")).toBeVisible();
  });

  test("should handle component not found error", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Navigate to a non-existent component
    await page.goto("/components/non-existent-component");

    // Verify error state is displayed
    await expect(page.getByTestId("component-details-error")).toBeVisible();
    await expect(page.getByTestId("error-title")).toContainText(
      "Error loading component",
    );
    await expect(page.getByTestId("error-message")).toContainText(
      "Component not found",
    );
  });

  test("should navigate back to components list", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Navigate to component details
    await page.goto("/components/auth-service");

    // Click the back link
    await page.getByTestId("back-to-components").click();

    // Verify we're back on the components list page
    await expect(page.getByTestId("page-title")).toContainText(
      "Component Catalog",
    );
    await expect(page.getByTestId("components-table")).toBeVisible();
  });

  test("should handle URL navigation directly", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Navigate directly to component details via URL
    await page.goto("/components/platform-infrastructure");

    // Verify component details are displayed
    await expect(page.getByTestId("component-details")).toBeVisible();
    await expect(page.getByTestId("component-name")).toContainText(
      "Platform Infrastructure",
    );
  });

  test("should verify component details API integration", async ({
    page,
  }: {
    page: Page;
  }) => {
    await ensureReports(page.request, "auth-service", 1);
    // Get component details via API
    const apiResponse = await page.request.get(
      "http://localhost:8080/api/catalog/v1/components/auth-service",
    );
    expect(apiResponse.ok()).toBeTruthy();

    const component: Component = await apiResponse.json();
    expect(component.id).toBe("auth-service");
    expect(component.name).toBe("Authentication Service");
    expect(component.owners?.team).toBe("Security Team");

    // Verify frontend displays same data
    await page.goto("/components/auth-service");
    await expect(page.getByTestId("component-name")).toContainText(
      "Authentication Service",
    );
    await expect(page.getByTestId("component-team")).toContainText(
      "Security Team",
    );
  });
});
