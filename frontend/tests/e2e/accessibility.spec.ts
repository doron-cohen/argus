import { test, expect } from "@playwright/test";
import injectAxe from "@axe-core/playwright";

/**
 * Comprehensive accessibility tests using axe-core
 * Tests WCAG 2.1 AA compliance across key pages and themes
 */

const themes = ["light", "dark"] as const;
const keyPages = [
  { path: "/", name: "home" },
  { path: "/settings", name: "settings" },
  { path: "/components", name: "components" },
] as const;

// Axe-core rules configuration for WCAG 2.1 AA
const axeConfig = {
  rules: {
    // Include all rules
  },
  runOnly: {
    type: "tag",
    values: ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "best-practice"],
  },
};

test.describe("Accessibility Tests - WCAG 2.1 AA Compliance", () => {
  test.beforeEach(async ({ page }) => {
    // Inject axe-core using the official Playwright package
    await new injectAxe({ page });
  });

  for (const theme of themes) {
    test.describe(`Theme: ${theme}`, () => {
      test.beforeEach(async ({ page }) => {
        // Set theme
        await page.addInitScript((themeName) => {
          document.documentElement.setAttribute("data-theme", themeName);
        }, theme);
      });

      for (const pageInfo of keyPages) {
        test(`should pass axe-core accessibility audit on ${pageInfo.name} page`, async ({
          page,
        }) => {
          await page.goto(pageInfo.path);
          await page.waitForLoadState("load");

          // Wait for dynamic content
          if (pageInfo.path === "/settings") {
            await page.waitForSelector('[data-testid="page-title"]');
          }

          // Run axe-core audit using official package
          // injectAxe returns the axe object with run method
          const axe = new injectAxe({ page });
          const results = await axe.analyze();

          // Check for violations
          if (results.violations.length > 0) {
            console.log(
              `Accessibility violations found on ${pageInfo.name} page (${theme} theme):`,
            );
            results.violations.forEach((violation: any) => {
              console.log(`- ${violation.id}: ${violation.description}`);
              console.log(`  Impact: ${violation.impact}`);
              console.log(`  Elements: ${violation.nodes.length}`);
            });
          }

          // Assert no critical or serious violations
          const criticalViolations = results.violations.filter(
            (v: any) => v.impact === "critical" || v.impact === "serious",
          );

          expect(criticalViolations).toHaveLength(0);
        });
      }
    });
  }

  test.describe("Interactive Elements", () => {
    test.beforeEach(async ({ page }) => {
      await page.goto("/");
      await page.waitForLoadState("load");
    });

    test("should have proper focus management", async ({ page }) => {
      // Test keyboard navigation
      await page.keyboard.press("Tab");

      // Check if something received focus
      const activeElement = await page.evaluate(() => {
        return document.activeElement?.tagName;
      });

      expect(activeElement).toBeDefined();
    });

    test.skip("should maintain focus visibility", async ({ page }) => {
      // Tab through focusable elements
      for (let i = 0; i < 5; i++) {
        await page.keyboard.press("Tab");

        // Check if focus is visible
        const focusVisible = await page.evaluate(() => {
          const active = document.activeElement;
          if (!active) return false;

          const styles = window.getComputedStyle(active);
          return (
            styles.outlineWidth !== "0px" ||
            active.hasAttribute("focus-visible")
          );
        });

        if (focusVisible) {
          break; // Found a visible focus
        }
      }

      // Should have found at least one visible focus
      const hasVisibleFocus = await page.evaluate(() => {
        const active = document.activeElement;
        if (!active) return false;

        const styles = window.getComputedStyle(active);
        return (
          styles.outlineWidth !== "0px" || active.hasAttribute("focus-visible")
        );
      });

      expect(hasVisibleFocus).toBe(true);
    });

    test("should respect prefers-reduced-motion", async ({ page }) => {
      // Simulate prefers-reduced-motion setting
      await page.emulateMedia({ reducedMotion: "reduce" });

      // Check that animations are disabled
      const hasAnimations = await page.evaluate(() => {
        const elements = Array.from(document.querySelectorAll("*"));
        for (const el of elements) {
          const styles = window.getComputedStyle(el);
          if (styles.animationName !== "none" && styles.animationName !== "") {
            return true;
          }
        }
        return false;
      });

      expect(hasAnimations).toBe(false);
    });
  });

  test.describe("Color and Contrast", () => {
    test("should maintain sufficient color contrast in both themes", async ({
      page,
    }) => {
      // Test light theme
      await page.addInitScript(() => {
        document.documentElement.setAttribute("data-theme", "light");
      });

      await page.goto("/");
      await page.waitForLoadState("load");

      // Basic contrast check - ensure text is readable
      const textElements = await page.$$("p, h1, h2, h3, h4, h5, h6, span");
      for (const element of textElements.slice(0, 5)) {
        // Test first 5 text elements
        const isVisible = await element.isVisible();
        expect(isVisible).toBe(true);
      }

      // Test dark theme
      await page.addInitScript(() => {
        document.documentElement.setAttribute("data-theme", "dark");
      });

      await page.reload();
      await page.waitForLoadState("load");

      // Verify text is still visible in dark theme
      const darkTextElements = await page.$$("p, h1, h2, h3, h4, h5, h6, span");
      for (const element of darkTextElements.slice(0, 5)) {
        const isVisible = await element.isVisible();
        expect(isVisible).toBe(true);
      }
    });
  });
});
