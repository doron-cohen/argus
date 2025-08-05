import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { ComponentDetails } from "./component-details";
import {
  componentDetails,
  loading,
  error,
  setComponentDetails,
  setLoading,
  setError,
  resetComponentDetails,
} from "../stores/app-store";
import type { Component } from "../stores/app-store";

// Mock data
const mockComponent: Component = {
  id: "test-component",
  name: "Test Component",
  description: "This is a test component",
  owners: {
    maintainers: ["john.doe", "jane.smith"],
    team: "Platform Team",
  },
};

const mockComponentWithoutOptionalFields: Component = {
  id: "minimal-component",
  name: "Minimal Component",
  description: "",
  owners: {
    maintainers: [],
    team: "",
  },
};

describe("ComponentDetails", () => {
  let element: ComponentDetails;

  beforeEach(() => {
    // Reset stores before each test
    resetComponentDetails();

    // Create fresh element
    element = new ComponentDetails();
    document.body.appendChild(element);
  });

  afterEach(() => {
    // Clean up stores
    resetComponentDetails();
  });

  describe("Component rendering", () => {
    test("should render component details when component data is provided", () => {
      setComponentDetails(mockComponent);

      const nameElement = element.querySelector(
        '[data-testid="component-name"]'
      );
      const idElement = element.querySelector('[data-testid="component-id"]');
      const descriptionElement = element.querySelector(
        '[data-testid="component-description"]'
      );
      const teamElement = element.querySelector(
        '[data-testid="component-team"]'
      );
      const maintainersElement = element.querySelector(
        '[data-testid="component-maintainers"]'
      );

      expect(nameElement?.textContent?.trim()).toBe("Test Component");
      expect(idElement?.textContent?.trim()).toBe("ID: test-component");
      expect(descriptionElement?.textContent?.trim()).toBe(
        "This is a test component"
      );
      expect(teamElement?.textContent?.trim()).toBe("Platform Team");
      expect(maintainersElement?.textContent?.trim()).toBe(
        "john.doe, jane.smith"
      );
    });

    test("should handle component with missing optional fields", () => {
      setComponentDetails(mockComponentWithoutOptionalFields);

      const descriptionElement = element.querySelector(
        '[data-testid="component-description"]'
      );
      const teamElement = element.querySelector(
        '[data-testid="component-team"]'
      );
      const maintainersElement = element.querySelector(
        '[data-testid="component-maintainers"]'
      );

      expect(descriptionElement?.textContent?.trim()).toBe(
        "No description available"
      );
      expect(teamElement?.textContent?.trim()).toBe("No team assigned");
      expect(maintainersElement?.textContent?.trim()).toBe(
        "No maintainers assigned"
      );
    });

    test("should use component name as ID when id field is missing", () => {
      const componentWithoutId = { ...mockComponent };
      delete (componentWithoutId as any).id;

      setComponentDetails(componentWithoutId);

      const idElement = element.querySelector('[data-testid="component-id"]');
      expect(idElement?.textContent?.trim()).toBe("ID: Test Component");
    });

    test("should render empty content when component is null", () => {
      setComponentDetails(null);

      expect(element.innerHTML).toBe("");
    });

    test("should include back to components link", () => {
      setComponentDetails(mockComponent);

      const backLink = element.querySelector(
        '[data-testid="back-to-components"]'
      ) as HTMLAnchorElement;
      expect(backLink).toBeTruthy();
      expect(backLink.href).toBe("/");
      expect(backLink.textContent?.trim()).toBe("← Back to Components");
    });
  });

  describe("Loading state", () => {
    test("should render loading skeleton when loading is true", () => {
      setLoading(true);

      const loadingElement = element.querySelector(
        '[data-testid="component-details-loading"]'
      );
      expect(loadingElement).toBeTruthy();

      // Check for loading skeleton elements
      const pulseElements = element.querySelectorAll(".animate-pulse");
      expect(pulseElements.length).toBeGreaterThan(0);

      const placeholders = element.querySelectorAll(".bg-gray-200");
      expect(placeholders.length).toBeGreaterThan(0);
    });

    test("should not show loading when loading is false", () => {
      setLoading(false);
      setComponentDetails(mockComponent);

      const loadingElement = element.querySelector(
        '[data-testid="component-details-loading"]'
      );
      expect(loadingElement).toBeFalsy();

      const componentElement = element.querySelector(
        '[data-testid="component-details"]'
      );
      expect(componentElement).toBeTruthy();
    });
  });

  describe("Error state", () => {
    test("should render error message when error is set", () => {
      const errorMessage = "Failed to load component details";
      setError(errorMessage);

      const errorElement = element.querySelector(
        '[data-testid="component-details-error"]'
      );
      expect(errorElement).toBeTruthy();

      const errorTitle = element.querySelector('[data-testid="error-title"]');
      expect(errorTitle?.textContent?.trim()).toBe("Error loading component");

      const errorMessageElement = element.querySelector(
        '[data-testid="error-message"]'
      );
      expect(errorMessageElement?.textContent?.trim()).toBe(errorMessage);
    });

    test("should include back link in error state", () => {
      setError("Some error");

      const backLink = element.querySelector(
        '[data-testid="back-to-components"]'
      ) as HTMLAnchorElement;
      expect(backLink).toBeTruthy();
      expect(backLink.href).toBe("/");
    });

    test("should not show error when error is null", () => {
      setError(null);
      setComponentDetails(mockComponent);

      const errorElement = element.querySelector(
        '[data-testid="component-details-error"]'
      );
      expect(errorElement).toBeFalsy();
    });
  });

  describe("XSS Protection", () => {
    test("should escape HTML in component data", () => {
      const maliciousComponent: Component = {
        id: '<script>alert("xss")</script>',
        name: '<img src="x" onerror="alert(1)">',
        description: '<script>console.log("hack")</script>',
        owners: {
          maintainers: ['<script>alert("maintainer")</script>'],
          team: '<script>alert("team")</script>',
        },
      };

      setComponentDetails(maliciousComponent);

      // Check that script tags are escaped and not executed
      expect(element.innerHTML).not.toContain("<script>");
      expect(element.innerHTML).toContain("&lt;script&gt;");

      // Verify actual displayed text is escaped
      const nameElement = element.querySelector(
        '[data-testid="component-name"]'
      );
      expect(nameElement?.textContent).toContain(
        '<img src="x" onerror="alert(1)">'
      );
    });

    test("should escape HTML in error messages", () => {
      const maliciousError = '<script>alert("error xss")</script>';
      setError(maliciousError);

      expect(element.innerHTML).not.toContain(
        '<script>alert("error xss")</script>'
      );
      expect(element.innerHTML).toContain("&lt;script&gt;");

      const errorMessageElement = element.querySelector(
        '[data-testid="error-message"]'
      );
      expect(errorMessageElement?.textContent?.trim()).toBe(maliciousError);
    });
  });

  describe("State management integration", () => {
    test("should react to store changes", () => {
      // Initial state
      expect(element.innerHTML).toBe("");

      // Set loading
      setLoading(true);
      expect(
        element.querySelector('[data-testid="component-details-loading"]')
      ).toBeTruthy();

      // Set component data (should override loading)
      setComponentDetails(mockComponent);
      expect(
        element.querySelector('[data-testid="component-details"]')
      ).toBeTruthy();
      expect(
        element.querySelector('[data-testid="component-details-loading"]')
      ).toBeFalsy();

      // Set error (should override component data)
      setError("Test error");
      expect(
        element.querySelector('[data-testid="component-details-error"]')
      ).toBeTruthy();
      expect(
        element.querySelector('[data-testid="component-details"]')
      ).toBeFalsy();
    });

    test("should clean up subscriptions on disconnect", () => {
      // Connect element and verify it's subscribed
      setComponentDetails(mockComponent);
      expect(
        element.querySelector('[data-testid="component-details"]')
      ).toBeTruthy();

      // Spy on the subscriptions array (through accessing private property for testing)
      const subscriptionCount = (element as any).subscriptions?.length;
      expect(subscriptionCount).toBeGreaterThan(0);

      // Disconnect element
      element.disconnectedCallback();

      // Verify subscriptions were cleaned up
      const subscriptionCountAfter = (element as any).subscriptions?.length;
      expect(subscriptionCountAfter).toBe(0);
    });
  });
});
