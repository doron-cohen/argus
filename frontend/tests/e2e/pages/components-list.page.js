import { expect } from "@playwright/test";

export class ComponentsListPage {
  constructor(page) {
    this.page = page;

    // Locators using data-testid attributes for stability
    this.pageTitle = page.getByTestId("page-title");
    this.pageDescription = page.getByTestId("page-description");
    this.searchInput = page.getByTestId("search-input");
    this.loadingSpinner = page.getByTestId("loading-spinner");
    this.errorMessage = page.getByTestId("error-message");
    this.componentsContainer = page.getByTestId("components-container");
    this.componentsHeader = page.getByTestId("components-header");
    this.componentsTable = page.getByTestId("components-table");
    this.componentsTbody = page.getByTestId("components-tbody");
    this.componentRows = page.getByTestId("component-row");
    this.emptyState = page.getByTestId("empty-state");
    this.emptyStateTitle = page.getByTestId("empty-state-title");
    this.emptyStateMessage = page.getByTestId("empty-state-message");

    // Table headers
    this.headerName = page.getByTestId("header-name");
    this.headerId = page.getByTestId("header-id");
    this.headerDescription = page.getByTestId("header-description");
    this.headerTeam = page.getByTestId("header-team");
    this.headerMaintainers = page.getByTestId("header-maintainers");
  }

  /**
   * Navigate to the components list page
   */
  async goto() {
    await this.page.goto("/");
  }

  /**
   * Wait for the page to load completely
   */
  async waitForPageLoad() {
    await this.pageTitle.waitFor({ state: "visible" });
    await this.loadingSpinner.waitFor({ state: "hidden" });
  }

  /**
   * Wait for components to be displayed
   */
  async waitForComponentsLoaded() {
    await this.componentsContainer.waitFor({ state: "visible" });
    await this.componentsTable.waitFor({ state: "visible" });
  }

  /**
   * Search for components using the search input
   * @param {string} query - The search query
   */
  async searchComponents(query) {
    await this.searchInput.fill(query);
    await this.searchInput.press("Enter");
  }

  /**
   * Clear the search input
   */
  async clearSearch() {
    await this.searchInput.clear();
    await this.searchInput.press("Enter");
  }

  /**
   * Get the number of displayed components
   * @returns {Promise<number>} The number of component rows
   */
  async getComponentCount() {
    return await this.componentRows.count();
  }

  /**
   * Get all component names
   * @returns {Promise<string[]>} Array of component names
   */
  async getComponentNames() {
    const names = [];
    const count = await this.componentRows.count();

    for (let i = 0; i < count; i++) {
      const name = await this.componentRows
        .nth(i)
        .getByTestId("component-name")
        .textContent();
      names.push(name);
    }

    return names;
  }

  /**
   * Get component by index
   * @param {number} index - The index of the component (0-based)
   * @returns {Promise<Object>} Component data
   */
  async getComponentByIndex(index) {
    const row = this.componentRows.nth(index);

    return {
      name: await row.getByTestId("component-name").textContent(),
      id: await row.getByTestId("component-id").textContent(),
      description: await row.getByTestId("component-description").textContent(),
      team: await row.getByTestId("component-team").textContent(),
      maintainers: await row.getByTestId("component-maintainers").textContent(),
    };
  }

  /**
   * Verify that the page displays the expected dummy components
   */
  async verifyDummyComponentsDisplayed() {
    await this.waitForComponentsLoaded();

    const expectedComponents = [
      "Authentication Service",
      "User Management Service",
      "Payment Processing Service",
    ];

    const actualComponents = await this.getComponentNames();

    expect(actualComponents).toEqual(expectedComponents);
    expect(await this.getComponentCount()).toBe(3);
  }

  /**
   * Verify that search functionality works correctly
   * @param {string} searchQuery - The search query
   * @param {string} expectedComponent - The expected component name
   */
  async verifySearchResults(searchQuery, expectedComponent) {
    await this.searchComponents(searchQuery);

    // Wait for search results to update
    await this.page.waitForTimeout(500);

    const componentNames = await this.getComponentNames();
    expect(componentNames).toContain(expectedComponent);
  }

  /**
   * Verify that empty state is displayed when no results found
   */
  async verifyEmptyStateDisplayed() {
    await this.emptyState.waitFor({ state: "visible" });
    await expect(this.emptyStateTitle).toContainText("No components found");
  }

  /**
   * Verify that all table headers are present
   */
  async verifyTableHeaders() {
    await expect(this.headerName).toContainText("Name");
    await expect(this.headerId).toContainText("ID");
    await expect(this.headerDescription).toContainText("Description");
    await expect(this.headerTeam).toContainText("Team");
    await expect(this.headerMaintainers).toContainText("Maintainers");
  }

  /**
   * Verify that loading state is displayed initially
   */
  async verifyLoadingState() {
    await this.loadingSpinner.waitFor({ state: "visible" });
  }
}
