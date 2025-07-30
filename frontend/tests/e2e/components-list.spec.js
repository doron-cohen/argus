import { test, expect } from "@playwright/test";
import { ComponentsListPage } from "./pages/components-list.page.js";
import { TestHelpers } from "./utils/test-helpers.js";

test.describe("Components List Page", () => {
  let componentsPage;
  let testHelpers;

  test.beforeEach(async ({ page }) => {
    componentsPage = new ComponentsListPage(page);
    testHelpers = new TestHelpers(page);
  });

  test("should display page title and description correctly", async ({
    page,
  }) => {
    // Given: I navigate to the components list page
    await componentsPage.goto();

    // When: The page loads completely
    await componentsPage.waitForPageLoad();

    // Then: I should see the correct page title and description
    await testHelpers.expectElementToContainText(
      componentsPage.pageTitle,
      "Component Catalog"
    );
    await testHelpers.expectElementToContainText(
      componentsPage.pageDescription,
      "Browse and search components in the Argus catalog"
    );
  });

  test("should show loading spinner initially and then display components", async ({
    page,
  }) => {
    // Given: I navigate to the components list page
    await componentsPage.goto();

    // When: The page starts loading
    // Then: I should see the loading spinner
    await componentsPage.verifyLoadingState();

    // When: The page finishes loading
    await componentsPage.waitForPageLoad();

    // Then: I should see the components table with dummy data
    await componentsPage.verifyDummyComponentsDisplayed();
  });

  test("should display all three dummy components with correct information", async ({
    page,
  }) => {
    // Given: I navigate to the components list page
    await componentsPage.goto();
    await componentsPage.waitForPageLoad();

    // When: The components are loaded
    await componentsPage.waitForComponentsLoaded();

    // Then: I should see all three dummy components
    await componentsPage.verifyDummyComponentsDisplayed();

    // And: I should see the correct component details
    const firstComponent = await componentsPage.getComponentByIndex(0);
    expect(firstComponent.name).toBe("Authentication Service");
    expect(firstComponent.id).toBe("auth-service");
    expect(firstComponent.team).toBe("Platform Team");
  });

  test("should display all required table headers", async ({ page }) => {
    // Given: I navigate to the components list page
    await componentsPage.goto();
    await componentsPage.waitForPageLoad();

    // When: The components table is displayed
    await componentsPage.waitForComponentsLoaded();

    // Then: I should see all the required table headers
    await componentsPage.verifyTableHeaders();
  });

  test("should filter components when searching by name", async ({ page }) => {
    // Given: I navigate to the components list page and wait for it to load
    await componentsPage.goto();
    await componentsPage.waitForPageLoad();
    await componentsPage.waitForComponentsLoaded();

    // When: I search for "auth"
    await componentsPage.searchComponents("auth");

    // Then: I should see only the Authentication Service
    await componentsPage.verifySearchResults("auth", "Authentication Service");
    expect(await componentsPage.getComponentCount()).toBe(1);
  });

  test("should filter components when searching by description", async ({
    page,
  }) => {
    // Given: I navigate to the components list page and wait for it to load
    await componentsPage.goto();
    await componentsPage.waitForPageLoad();
    await componentsPage.waitForComponentsLoaded();

    // When: I search for "management"
    await componentsPage.searchComponents("management");

    // Then: I should see only the User Management Service
    await componentsPage.verifySearchResults(
      "management",
      "User Management Service"
    );
    expect(await componentsPage.getComponentCount()).toBe(1);
  });

  test("should perform case-insensitive search", async ({ page }) => {
    // Given: I navigate to the components list page and wait for it to load
    await componentsPage.goto();
    await componentsPage.waitForPageLoad();
    await componentsPage.waitForComponentsLoaded();

    // When: I search for "SERVICE" (uppercase)
    await componentsPage.searchComponents("SERVICE");

    // Then: I should see all three components (case-insensitive search)
    const componentNames = await componentsPage.getComponentNames();
    expect(componentNames).toHaveLength(3);
    expect(componentNames).toContain("Authentication Service");
    expect(componentNames).toContain("User Management Service");
    expect(componentNames).toContain("Payment Processing Service");
  });

  test("should show all components when search is cleared", async ({
    page,
  }) => {
    // Given: I navigate to the components list page and wait for it to load
    await componentsPage.goto();
    await componentsPage.waitForPageLoad();
    await componentsPage.waitForComponentsLoaded();

    // When: I search for something that doesn't exist
    await componentsPage.searchComponents("nonexistent");

    // And: I clear the search
    await componentsPage.clearSearch();

    // Then: I should see all three components again
    await componentsPage.verifyDummyComponentsDisplayed();
  });

  test("should display empty state when no search results are found", async ({
    page,
  }) => {
    // Given: I navigate to the components list page and wait for it to load
    await componentsPage.goto();
    await componentsPage.waitForPageLoad();
    await componentsPage.waitForComponentsLoaded();

    // When: I search for something that doesn't exist
    await componentsPage.searchComponents("nonexistent-component");

    // Then: I should see the empty state
    await componentsPage.verifyEmptyStateDisplayed();

    // And: The empty state should have the correct message
    await testHelpers.expectElementToContainText(
      componentsPage.emptyStateMessage,
      "Try adjusting your search terms."
    );
  });

  test("should maintain search functionality across multiple searches", async ({
    page,
  }) => {
    // Given: I navigate to the components list page and wait for it to load
    await componentsPage.goto();
    await componentsPage.waitForPageLoad();
    await componentsPage.waitForComponentsLoaded();

    // When: I search for "auth"
    await componentsPage.searchComponents("auth");
    expect(await componentsPage.getComponentCount()).toBe(1);

    // And: I search for "payment"
    await componentsPage.searchComponents("payment");
    expect(await componentsPage.getComponentCount()).toBe(1);

    // And: I search for "user"
    await componentsPage.searchComponents("user");
    expect(await componentsPage.getComponentCount()).toBe(1);

    // Then: Each search should return the correct results
    const componentNames = await componentsPage.getComponentNames();
    expect(componentNames).toContain("User Management Service");
  });
});
