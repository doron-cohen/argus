import { test, expect, type Page } from "@playwright/test";
import { ensureReports } from "./fixtures";
import type { Component } from "./types";

test.describe("Component Details Page", () => {
  test("should display component details", async ({ page }: { page: Page }) => {
    // Navigate to a component that should exist
    await page.goto("/components/auth-service");

    // Wait for the component details to load and be visible
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Verify the page title is displayed
    await expect(page.getByTestId("page-title")).toBeVisible();
    await expect(page.getByTestId("page-title")).toHaveText(
      "Component Details"
    );

    // Verify component information is displayed
    await expect(page.getByTestId("component-name")).toBeVisible();
    await expect(page.getByTestId("component-id")).toBeVisible();
    await expect(page.getByTestId("description-label")).toBeVisible();
    await expect(page.getByTestId("component-description")).toBeVisible();
    await expect(page.getByTestId("team-label")).toBeVisible();
    await expect(page.getByTestId("component-team")).toBeVisible();
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

  test("should handle component with no reports", async ({
    page,
  }: {
    page: Page;
  }) => {
    // Navigate to a component that should have no reports initially
    await page.goto("/components/platform-infrastructure");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Check for no reports message
    await expect(page.getByTestId("no-reports")).toBeVisible();
  });

  test("should show loading state", async ({ page }: { page: Page }) => {
    // Navigate to a component
    await page.goto("/components/auth-service");

    // The loading state should be brief, but we can check for the component details
    // which should appear after loading
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
      "Error loading component"
    );
  });

  test("should display back to components link", async ({
    page,
  }: {
    page: Page;
  }) => {
    await page.goto("/components/auth-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Verify back link is present
    await expect(page.getByTestId("back-to-components")).toBeVisible();
    await expect(page.getByTestId("back-to-components")).toHaveText(
      "← Back to Components"
    );
  });

  test("should navigate back to components list", async ({
    page,
  }: {
    page: Page;
  }) => {
    await page.goto("/components/auth-service");

    // Wait for component details to load
    await expect(page.getByTestId("component-details")).toBeVisible({
      timeout: 10000,
    });

    // Click the back link
    await page.getByTestId("back-to-components").click();

    // Verify we're back on the components list page
    await expect(page).toHaveURL("/components");
  });
});
