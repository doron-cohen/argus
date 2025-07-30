import { test, expect, describe, beforeEach } from "bun:test";
import ComponentsList from "../../src/components/components-list.js";

describe("ComponentsList", () => {
  let component;

  beforeEach(() => {
    component = ComponentsList();
  });

  test("loads dummy components on init", async () => {
    await component.init();

    expect(component.components).toHaveLength(3);
    expect(component.loading).toBe(false);
    expect(component.error).toBe(null);

    // Check that dummy components are loaded
    expect(component.components[0].name).toBe("Authentication Service");
    expect(component.components[1].name).toBe("User Management Service");
    expect(component.components[2].name).toBe("Payment Processing Service");
  });

  test("filters components by search query", () => {
    // Test name search
    component.searchQuery = "auth";
    expect(component.filteredComponents).toHaveLength(1);
    expect(component.filteredComponents[0].name).toBe("Authentication Service");

    // Test description search
    component.searchQuery = "management";
    expect(component.filteredComponents).toHaveLength(1);
    expect(component.filteredComponents[0].name).toBe(
      "User Management Service"
    );

    // Test case insensitive search
    component.searchQuery = "SERVICE";
    expect(component.filteredComponents).toHaveLength(3);

    // Test empty search
    component.searchQuery = "";
    expect(component.filteredComponents).toHaveLength(3);
  });

  test("truncates text correctly", () => {
    const longText =
      "This is a very long description that should be truncated when it exceeds the maximum length";
    const shortText = "Short text";

    expect(component.truncateText(longText, 50)).toBe(
      "This is a very long description that should be ..."
    );
    expect(component.truncateText(shortText, 50)).toBe(shortText);
    expect(component.truncateText("", 50)).toBe("");
    expect(component.truncateText(null, 50)).toBe("");
  });

  test("formats date correctly", () => {
    const dateString = "2023-12-01T10:30:00Z";
    const formatted = component.formatDate(dateString);

    expect(formatted).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}/); // Basic date format check

    expect(component.formatDate("")).toBe("N/A");
    expect(component.formatDate(null)).toBe("N/A");
  });

  test("dummy components have correct structure", () => {
    const components = component.components;

    expect(components).toHaveLength(3);

    // Check first component structure
    expect(components[0]).toHaveProperty("id", "auth-service");
    expect(components[0]).toHaveProperty("name", "Authentication Service");
    expect(components[0]).toHaveProperty("description");
    expect(components[0]).toHaveProperty("owners");
    expect(components[0].owners).toHaveProperty("team");
    expect(components[0].owners).toHaveProperty("maintainers");
    expect(Array.isArray(components[0].owners.maintainers)).toBe(true);
  });
});
