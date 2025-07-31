import { test, expect } from "@playwright/test";

test.describe("Component Catalog", () => {
  test("should display page title and description", async ({ page }) => {
    await page.goto("/");

    // Check page title and description
    await expect(page.getByTestId("page-title")).toContainText(
      "Component Catalog"
    );
    await expect(page.getByTestId("page-description")).toContainText(
      "Browse and search components in the Argus catalog"
    );
  });

  test("should display components table with headers", async ({ page }) => {
    await page.goto("/");

    // Check table headers
    await expect(page.getByTestId("header-name")).toContainText("Name");
    await expect(page.getByTestId("header-id")).toContainText("ID");
    await expect(page.getByTestId("header-description")).toContainText(
      "Description"
    );
    await expect(page.getByTestId("header-team")).toContainText("Team");
    await expect(page.getByTestId("header-maintainers")).toContainText(
      "Maintainers"
    );
  });

  test("should display all three components", async ({ page }) => {
    await page.goto("/");

    // Check that we have 3 component rows
    const componentRows = page.getByTestId("component-row");
    await expect(componentRows).toHaveCount(3);

    // Check first component details
    const firstComponent = componentRows.first();
    await expect(firstComponent.getByTestId("component-name")).toContainText(
      "Authentication Service"
    );
    await expect(firstComponent.getByTestId("component-id")).toContainText(
      "auth-service"
    );
    await expect(firstComponent.getByTestId("component-team")).toContainText(
      "Security Team"
    );
  });

  test("should display component count correctly", async ({ page }) => {
    await page.goto("/");

    // Check that the component count shows 3
    await expect(page.getByTestId("components-header")).toContainText(
      "Components (3)"
    );
  });
});
