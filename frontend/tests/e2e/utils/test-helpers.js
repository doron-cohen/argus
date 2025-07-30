import { expect } from "@playwright/test";

/**
 * Common test utilities following Playwright best practices
 */
export class TestHelpers {
  constructor(page) {
    this.page = page;
  }

  /**
   * Wait for network idle to ensure all requests are complete
   */
  async waitForNetworkIdle() {
    await this.page.waitForLoadState("networkidle");
  }

  /**
   * Wait for a specific element to be visible and stable
   * @param {Locator} locator - The element locator
   * @param {number} timeout - Timeout in milliseconds
   */
  async waitForElementStable(locator, timeout = 5000) {
    await locator.waitFor({ state: "visible", timeout });

    // Wait a bit more to ensure the element is stable
    await this.page.waitForTimeout(100);
  }

  /**
   * Take a screenshot for debugging purposes
   * @param {string} name - Screenshot name
   */
  async takeScreenshot(name) {
    await this.page.screenshot({
      path: `test-results/${name}-${Date.now()}.png`,
      fullPage: true,
    });
  }

  /**
   * Verify that an element contains the expected text
   * @param {Locator} locator - The element locator
   * @param {string} expectedText - Expected text content
   */
  async expectElementToContainText(locator, expectedText) {
    await expect(locator).toContainText(expectedText);
  }

  /**
   * Verify that an element is visible
   * @param {Locator} locator - The element locator
   */
  async expectElementToBeVisible(locator) {
    await expect(locator).toBeVisible();
  }

  /**
   * Verify that an element is not visible
   * @param {Locator} locator - The element locator
   */
  async expectElementToBeHidden(locator) {
    await expect(locator).not.toBeVisible();
  }

  /**
   * Verify that the page has the expected title
   * @param {string} expectedTitle - Expected page title
   */
  async expectPageTitle(expectedTitle) {
    await expect(this.page).toHaveTitle(expectedTitle);
  }

  /**
   * Verify that the page URL matches the expected path
   * @param {string} expectedPath - Expected URL path
   */
  async expectPageUrl(expectedPath) {
    await expect(this.page).toHaveURL(expectedPath);
  }

  /**
   * Wait for a specific condition to be true
   * @param {Function} condition - Function that returns a boolean
   * @param {number} timeout - Timeout in milliseconds
   */
  async waitForCondition(condition, timeout = 5000) {
    const startTime = Date.now();

    while (Date.now() - startTime < timeout) {
      if (await condition()) {
        return;
      }
      await this.page.waitForTimeout(100);
    }

    throw new Error(`Condition not met within ${timeout}ms`);
  }

  /**
   * Retry an action with exponential backoff
   * @param {Function} action - The action to retry
   * @param {number} maxRetries - Maximum number of retries
   */
  async retryAction(action, maxRetries = 3) {
    let lastError;

    for (let i = 0; i < maxRetries; i++) {
      try {
        return await action();
      } catch (error) {
        lastError = error;

        if (i < maxRetries - 1) {
          // Exponential backoff: 100ms, 200ms, 400ms
          await this.page.waitForTimeout(100 * Math.pow(2, i));
        }
      }
    }

    throw lastError;
  }
}
