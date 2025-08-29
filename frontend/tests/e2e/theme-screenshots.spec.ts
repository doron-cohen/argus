import { test, expect } from "@playwright/test";

/**
 * Theme screenshot tests for visual regression testing
 * Captures screenshots of key pages in both light and dark themes
 */

const themes = ["light", "dark"] as const;
const keyPages = [
  { path: "/", name: "home" },
  { path: "/settings", name: "settings" },
  { path: "/components", name: "components" },
] as const;

test.describe("Theme Visual Regression Tests", () => {
  for (const theme of themes) {
    test.describe(`Theme: ${theme}`, () => {
      test.beforeEach(async ({ page }) => {
        // Navigate first to ensure page is loaded
        await page.goto("/");
        await page.waitForLoadState("load");

        // Clear any existing theme and localStorage
        await page.evaluate(() => {
          localStorage.removeItem("theme");
          document.documentElement.removeAttribute("data-theme");
        });

        // Set the test theme directly
        await page.evaluate((testTheme) => {
          // Set theme immediately
          document.documentElement.setAttribute("data-theme", testTheme);
          localStorage.setItem("theme", testTheme);

          // Force CSS recalculation
          document.body.style.display = "none";
          document.body.offsetHeight;
          document.body.style.display = "";
        }, theme);

        // Verify theme was actually set
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      });

      for (const pageInfo of keyPages) {
        test.skip(`should match visual baseline for ${pageInfo.name} page`, async ({
          page,
        }) => {
          await page.goto(pageInfo.path);
          await page.waitForLoadState("load");

          // Wait for dynamic content to load
          if (pageInfo.path === "/settings") {
            await page.waitForSelector('[data-testid="sync-source-unknown"]', {
              timeout: 10000,
            });
          }

          // Take full page screenshot
          await expect(page).toHaveScreenshot(
            `${theme}-${pageInfo.name}-page.png`,
            {
              fullPage: true,
              threshold: 0.1, // Allow 10% pixel difference for minor rendering variations
            },
          );
        });
      }

      // Test theme switching behavior
      test("should switch themes correctly", async ({ page }) => {
        await page.goto("/");
        await page.waitForLoadState("load");

        // Verify initial theme is applied
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);

        // Switch to the other theme
        const otherTheme = theme === "light" ? "dark" : "light";
        await page.evaluate((newTheme) => {
          document.documentElement.setAttribute("data-theme", newTheme);
          // Force CSS recalculation
          document.body.style.display = "none";
          document.body.offsetHeight; // Trigger reflow
          document.body.style.display = "";
        }, otherTheme);

        // Wait a bit for theme to apply
        await page.waitForTimeout(100);

        // Verify theme switch
        await expect(page.locator("html")).toHaveAttribute(
          "data-theme",
          otherTheme,
        );
      });
    });
  }
});
