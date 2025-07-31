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

  test("should show loading state initially", async ({ page }) => {
    // Set up the route mock before navigating to the page
    await page.route("/api/catalog/v1/components", async (route) => {
      // Add a longer delay to ensure the loading state is visible
      await new Promise((resolve) => setTimeout(resolve, 2000));
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: "auth-service",
            name: "Authentication Service",
            description:
              "Handles user authentication, authorization, and session management.",
            owners: {
              team: "Security Team",
              maintainers: ["alice.smith", "bob.jones"],
            },
          },
        ]),
      });
    });

    // Navigate to the page
    await page.goto("/");

    // Wait for the loading message to appear and check it
    await expect(page.getByTestId("loading-message")).toBeVisible();
    await expect(page.getByTestId("loading-message")).toContainText(
      "Loading components..."
    );
  });

  test("should display components from API", async ({ page }) => {
    // Mock the API response
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: "auth-service",
            name: "Authentication Service",
            description:
              "Handles user authentication, authorization, and session management.",
            owners: {
              team: "Security Team",
              maintainers: ["alice.smith", "bob.jones"],
            },
          },
          {
            id: "user-management",
            name: "User Management Service",
            description: "Manages user profiles, roles, and permissions.",
            owners: {
              team: "Platform Team",
              maintainers: ["carol.wilson", "dave.brown"],
            },
          },
        ]),
      });
    });

    await page.goto("/");

    // Wait for components to load
    await expect(page.getByTestId("component-row")).toHaveCount(2);

    // Check first component details
    const firstComponent = page.getByTestId("component-row").first();
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
    // Mock the API response
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: "auth-service",
            name: "Authentication Service",
            description:
              "Handles user authentication, authorization, and session management.",
            owners: {
              team: "Security Team",
              maintainers: ["alice.smith", "bob.jones"],
            },
          },
        ]),
      });
    });

    await page.goto("/");

    // Wait for components to load and check count
    await expect(page.getByTestId("components-header")).toContainText(
      "Components (1)"
    );
  });

  test("should handle empty components list", async ({ page }) => {
    // Mock empty API response
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([]),
      });
    });

    await page.goto("/");

    // Check that no components message is displayed
    await expect(page.getByTestId("no-components-message")).toContainText(
      "No components found"
    );
    await expect(page.getByTestId("components-header")).toContainText(
      "Components (0)"
    );
  });

  test("should handle API errors gracefully", async ({ page }) => {
    // Mock API error
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({
          error: "Internal server error",
          code: "INTERNAL_ERROR",
        }),
      });
    });

    await page.goto("/");

    // Check that error message is displayed
    await expect(page.getByTestId("error-message")).toContainText(
      "Error: Internal server error"
    );
  });
});
