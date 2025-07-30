import { expect } from "@playwright/test";

export async function searchComponents(page, query) {
  await page.fill("#search", query);
  await page.keyboard.press("Enter");
}

export async function clearSearch(page) {
  await page.fill("#search", "");
  await page.keyboard.press("Enter");
}

export async function waitForSearchResults(page, expectedCount) {
  await page.waitForFunction(
    (count) => document.querySelectorAll("tbody tr").length === count,
    expectedCount
  );
}
