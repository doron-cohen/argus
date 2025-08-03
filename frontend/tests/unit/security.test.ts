import { test, expect } from "@playwright/test";

// Mock the fetch function to test XSS prevention
const mockFetch = (response: any) => {
  return Promise.resolve({
    ok: true,
    json: () => Promise.resolve(response),
  });
};

// Test the escapeHtml function directly
function escapeHtml(unsafe: string): string {
  if (unsafe == null) return String(unsafe);
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

test.describe("XSS Prevention Unit Tests", () => {
  test("should escape HTML characters correctly", () => {
    const testCases = [
      {
        input: "<script>alert('XSS')</script>",
        expected: "&lt;script&gt;alert(&#039;XSS&#039;)&lt;/script&gt;",
      },
      {
        input: "<img src=x onerror=alert('XSS')>",
        expected: "&lt;img src=x onerror=alert(&#039;XSS&#039;)&gt;",
      },
      {
        input: "& < > \" '",
        expected: "&amp; &lt; &gt; &quot; &#039;",
      },
      {
        input: "Normal text",
        expected: "Normal text",
      },
      {
        input: "",
        expected: "",
      },
    ];

    testCases.forEach(({ input, expected }) => {
      expect(escapeHtml(input)).toBe(expected);
    });
  });

  test("should prevent XSS in component name", async ({ page }) => {
    // Mock the API response with malicious content
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: "xss-test",
            name: "<script>alert('XSS')</script>",
            description: "Test component",
            owners: {
              team: "security",
              maintainers: ["test"],
            },
          },
        ]),
      });
    });

    await page.goto("/");

    // Wait for component to load
    await page.waitForSelector('[data-testid="component-name"]', {
      timeout: 5000,
    });

    // Check that the script tag is escaped and not executed
    const nameElement = page.getByTestId("component-name");
    await expect(nameElement).toHaveText(
      "&lt;script&gt;alert(&#039;XSS&#039;)&lt;/script&gt;"
    );

    // Verify no actual script tags are present
    const scripts = page.locator("script");
    const scriptCount = await scripts.count();
    expect(scriptCount).toBeLessThanOrEqual(2); // Only legitimate app scripts
  });

  test("should prevent XSS in component description", async ({ page }) => {
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: "xss-test",
            name: "XSS Test",
            description: "<img src=x onerror=alert('XSS')>",
            owners: {
              team: "security",
              maintainers: ["test"],
            },
          },
        ]),
      });
    });

    await page.goto("/");

    await page.waitForSelector('[data-testid="component-description"]', {
      timeout: 5000,
    });

    const descElement = page.getByTestId("component-description");
    await expect(descElement).toHaveText(
      "&lt;img src=x onerror=alert(&#039;XSS&#039;)&gt;"
    );

    // Verify no img tags are present
    const images = page.locator("img");
    const imageCount = await images.count();
    expect(imageCount).toBe(0);
  });

  test("should prevent XSS in team and maintainers", async ({ page }) => {
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: "xss-test",
            name: "XSS Test",
            description: "Test component",
            owners: {
              team: "<script>alert('team')</script>",
              maintainers: ["<script>alert('maintainer')</script>"],
            },
          },
        ]),
      });
    });

    await page.goto("/");

    await page.waitForSelector('[data-testid="component-team"]', {
      timeout: 5000,
    });

    const teamElement = page.getByTestId("component-team");
    const maintainersElement = page.getByTestId("component-maintainers");

    await expect(teamElement).toHaveText(
      "&lt;script&gt;alert(&#039;team&#039;)&lt;/script&gt;"
    );
    await expect(maintainersElement).toHaveText(
      "&lt;script&gt;alert(&#039;maintainer&#039;)&lt;/script&gt;"
    );

    // Verify no script tags are present
    const scripts = page.locator("script");
    const scriptCount = await scripts.count();
    expect(scriptCount).toBeLessThanOrEqual(2);
  });

  test("should prevent XSS in error messages", async ({ page }) => {
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({
          error: "<script>alert('XSS error')</script>",
        }),
      });
    });

    await page.goto("/");

    await page.waitForSelector('[data-testid="error-message"]', {
      timeout: 5000,
    });

    const errorElement = page.getByTestId("error-message");
    await expect(errorElement).toContainText(
      "&lt;script&gt;alert(&#039;XSS error&#039;)&lt;/script&gt;"
    );

    // Verify no script tags are present
    const scripts = page.locator("script");
    const scriptCount = await scripts.count();
    expect(scriptCount).toBeLessThanOrEqual(2);
  });

  test("should handle multiple XSS vectors in single component", async ({
    page,
  }) => {
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: "<script>alert('id')</script>",
            name: "<script>alert('name')</script>",
            description: "<script>alert('desc')</script>",
            owners: {
              team: "<script>alert('team')</script>",
              maintainers: [
                "<script>alert('m1')</script>",
                "<script>alert('m2')</script>",
              ],
            },
          },
        ]),
      });
    });

    await page.goto("/");

    await page.waitForSelector('[data-testid="component-name"]', {
      timeout: 5000,
    });

    // Check all fields are escaped
    await expect(page.getByTestId("component-name")).toHaveText(
      "&lt;script&gt;alert(&#039;name&#039;)&lt;/script&gt;"
    );
    await expect(page.getByTestId("component-id")).toHaveText(
      "&lt;script&gt;alert(&#039;id&#039;)&lt;/script&gt;"
    );
    await expect(page.getByTestId("component-description")).toHaveText(
      "&lt;script&gt;alert(&#039;desc&#039;)&lt;/script&gt;"
    );
    await expect(page.getByTestId("component-team")).toHaveText(
      "&lt;script&gt;alert(&#039;team&#039;)&lt;/script&gt;"
    );
    await expect(page.getByTestId("component-maintainers")).toHaveText(
      "&lt;script&gt;alert(&#039;m1&#039;)&lt;/script&gt;, &lt;script&gt;alert(&#039;m2&#039;)&lt;/script&gt;"
    );

    // Verify no script execution
    const scripts = page.locator("script");
    const scriptCount = await scripts.count();
    expect(scriptCount).toBeLessThanOrEqual(2);
  });

  test("should handle edge cases and special characters", async ({ page }) => {
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: "edge-test",
            name: "& < > \" ' &amp; &lt; &gt; &quot; &#039;",
            description: "Test with already escaped content",
            owners: {
              team: "Team with & < > \" ' chars",
              maintainers: ["User with & < > \" ' chars"],
            },
          },
        ]),
      });
    });

    await page.goto("/");

    await page.waitForSelector('[data-testid="component-name"]', {
      timeout: 5000,
    });

    // Check that content is properly escaped
    await expect(page.getByTestId("component-name")).toHaveText(
      "&amp; &lt; &gt; &quot; &#039; &amp;amp; &amp;lt; &amp;gt; &amp;quot; &amp;#039;"
    );
    await expect(page.getByTestId("component-team")).toHaveText(
      "Team with &amp; &lt; &gt; &quot; &#039; chars"
    );
    await expect(page.getByTestId("component-maintainers")).toHaveText(
      "User with &amp; &lt; &gt; &quot; &#039; chars"
    );
  });

  test("should handle null and undefined values gracefully", async ({
    page,
  }) => {
    await page.route("/api/catalog/v1/components", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: null,
            name: undefined,
            description: null,
            owners: {
              team: null,
              maintainers: [null, undefined, ""],
            },
          },
        ]),
      });
    });

    await page.goto("/");

    await page.waitForSelector('[data-testid="component-row"]', {
      timeout: 5000,
    });

    // Check that null/undefined values are handled gracefully
    const row = page.getByTestId("component-row");
    await expect(row.getByTestId("component-name")).toHaveText("");
    await expect(row.getByTestId("component-id")).toHaveText("");
    await expect(row.getByTestId("component-description")).toHaveText("");
    await expect(row.getByTestId("component-team")).toHaveText("");
    await expect(row.getByTestId("component-maintainers")).toHaveText("");
  });
});
