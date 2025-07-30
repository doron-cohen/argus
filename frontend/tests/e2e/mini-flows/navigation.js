import { expect } from "@playwright/test";

export async function navigateToComponents(page) {
  await page.goto("/");
  await expect(page.locator("h1")).toContainText("Component Catalog");
}

export async function waitForComponentsLoad(page) {
  await page.waitForSelector('[x-data="componentsList"]');
  await expect(page.locator('[x-show="loading"]')).not.toBeVisible();
}

export async function waitForComponentsTable(page) {
  await page.waitForSelector("table.component-table");
  await expect(page.locator("table")).toBeVisible();
}
