import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { ComponentDetails } from "./component-details";
import {
  componentDetails,
  loading,
  error,
  latestReports,
  reportsLoading,
  reportsError,
  setComponentDetails,
  setLoading,
  setError,
  setLatestReports,
  setReportsLoading,
  setReportsError,
  resetComponentDetails,
  resetReports,
} from "../stores/app-store";
import type { Component, CheckReport } from "../stores/app-store";

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

const mockReports: CheckReport[] = [
  {
    id: "report-1",
    check_slug: "unit-tests",
    status: "pass",
    timestamp: "2024-01-15T10:30:00Z",
  },
  {
    id: "report-2",
    check_slug: "security-scan",
    status: "fail",
    timestamp: "2024-01-15T10:35:00Z",
  },
  {
    id: "report-3",
    check_slug: "code-quality",
    status: "disabled",
    timestamp: "2024-01-15T10:40:00Z",
  },
];

const mockEmptyReports: CheckReport[] = [];

describe("ComponentDetails", () => {
  let element: ComponentDetails;

  beforeEach(() => {
    // Reset stores before each test
    resetComponentDetails();
    resetReports();

    // Create fresh element
    element = new ComponentDetails();
    document.body.appendChild(element);
  });

  afterEach(() => {
    // Clean up stores
    resetComponentDetails();
    resetReports();
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

      // Since the component now sets textContent directly, ensure no script tags were injected
      expect(element.querySelectorAll("script").length).toBe(0);

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

      // Ensure no script tags exist in the DOM
      expect(element.querySelectorAll("script").length).toBe(0);

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

  describe("Reports functionality", () => {
    beforeEach(() => {
      setComponentDetails(mockComponent);
    });

    test("should render reports when reports data is provided", () => {
      setLatestReports(mockReports);

      const reportsContainer = element.querySelector("#reports-container");
      expect(reportsContainer).toBeTruthy();

      const reportsList = element.querySelector('[data-testid="reports-list"]');
      expect(reportsList).toBeTruthy();

      const reportItems = element.querySelectorAll(
        '[data-testid="report-item"]'
      );
      expect(reportItems.length).toBe(3);

      // Check first report (pass status)
      const firstReport = reportItems[0];
      const firstName = firstReport.querySelector('[data-testid="check-name"]');
      const firstStatus = firstReport.querySelector(
        '[data-testid="check-status"]'
      );
      expect(firstName?.textContent?.trim()).toBe("unit-tests");
      expect(firstStatus?.textContent?.trim()).toBe("pass");
      expect(firstStatus?.className).toContain("bg-green-100");

      // Check second report (fail status)
      const secondReport = reportItems[1];
      const secondName = secondReport.querySelector(
        '[data-testid="check-name"]'
      );
      const secondStatus = secondReport.querySelector(
        '[data-testid="check-status"]'
      );
      expect(secondName?.textContent?.trim()).toBe("security-scan");
      expect(secondStatus?.textContent?.trim()).toBe("fail");
      expect(secondStatus?.className).toContain("bg-red-100");

      // Check third report (disabled status)
      const thirdReport = reportItems[2];
      const thirdName = thirdReport.querySelector('[data-testid="check-name"]');
      const thirdStatus = thirdReport.querySelector(
        '[data-testid="check-status"]'
      );
      expect(thirdName?.textContent?.trim()).toBe("code-quality");
      expect(thirdStatus?.textContent?.trim()).toBe("disabled");
      expect(thirdStatus?.className).toContain("bg-yellow-100");
    });

    test("should render empty state when no reports are available", () => {
      setLatestReports(mockEmptyReports);

      const reportsContainer = element.querySelector("#reports-container");
      expect(reportsContainer).toBeTruthy();

      const noReports = element.querySelector('[data-testid="no-reports"]');
      expect(noReports).toBeTruthy();
      expect(noReports?.textContent?.trim()).toBe(
        "No quality checks available"
      );

      const reportsList = element.querySelector('[data-testid="reports-list"]');
      expect(reportsList).toBeFalsy();
    });

    test("should show reports loading state", () => {
      setReportsLoading(true);

      const reportsContainer = element.querySelector("#reports-container");
      expect(reportsContainer).toBeTruthy();

      const loadingElement = element.querySelector(
        '[data-testid="reports-loading"]'
      );
      expect(loadingElement).toBeTruthy();
      expect(loadingElement?.textContent?.trim()).toBe(
        "Loading quality checks..."
      );
    });

    test("should show reports error state", () => {
      const errorMessage = "Failed to load reports";
      setReportsError(errorMessage);

      const reportsContainer = element.querySelector("#reports-container");
      expect(reportsContainer).toBeTruthy();

      const errorElement = element.querySelector(
        '[data-testid="reports-error"]'
      );
      expect(errorElement).toBeTruthy();
      expect(errorElement?.textContent?.trim()).toBe(
        "Error loading quality checks: " + errorMessage
      );
    });

    test("should display timestamps correctly", () => {
      setLatestReports(mockReports);

      const timestamps = element.querySelectorAll(
        '[data-testid="check-timestamp"]'
      );
      expect(timestamps.length).toBe(3);

      // Check that timestamps are formatted (should contain date and time)
      const firstTimestamp = timestamps[0]?.textContent?.trim();
      expect(firstTimestamp).toBeTruthy();
      expect(firstTimestamp).toContain("2024");
    });

    test("should handle all status types with correct colors and icons", () => {
      const allStatusReports: CheckReport[] = [
        {
          id: "1",
          check_slug: "test-pass",
          status: "pass",
          timestamp: "2024-01-15T10:30:00Z",
        },
        {
          id: "2",
          check_slug: "test-fail",
          status: "fail",
          timestamp: "2024-01-15T10:30:00Z",
        },
        {
          id: "3",
          check_slug: "test-error",
          status: "error",
          timestamp: "2024-01-15T10:30:00Z",
        },
        {
          id: "4",
          check_slug: "test-unknown",
          status: "unknown",
          timestamp: "2024-01-15T10:30:00Z",
        },
        {
          id: "5",
          check_slug: "test-disabled",
          status: "disabled",
          timestamp: "2024-01-15T10:30:00Z",
        },
        {
          id: "6",
          check_slug: "test-skipped",
          status: "skipped",
          timestamp: "2024-01-15T10:30:00Z",
        },
        {
          id: "7",
          check_slug: "test-completed",
          status: "completed",
          timestamp: "2024-01-15T10:30:00Z",
        },
      ];

      setLatestReports(allStatusReports);

      const reportItems = element.querySelectorAll(
        '[data-testid="report-item"]'
      );
      expect(reportItems.length).toBe(7);

      // Check pass status (green)
      const passStatus = reportItems[0].querySelector(
        '[data-testid="check-status"]'
      );
      expect(passStatus?.className).toContain("bg-green-100");
      expect(passStatus?.innerHTML).toContain("svg");

      // Check fail status (red)
      const failStatus = reportItems[1].querySelector(
        '[data-testid="check-status"]'
      );
      expect(failStatus?.className).toContain("bg-red-100");
      expect(failStatus?.innerHTML).toContain("svg");

      // Check error status (red)
      const errorStatus = reportItems[2].querySelector(
        '[data-testid="check-status"]'
      );
      expect(errorStatus?.className).toContain("bg-red-100");

      // Check unknown status (red)
      const unknownStatus = reportItems[3].querySelector(
        '[data-testid="check-status"]'
      );
      expect(unknownStatus?.className).toContain("bg-red-100");

      // Check disabled status (yellow)
      const disabledStatus = reportItems[4].querySelector(
        '[data-testid="check-status"]'
      );
      expect(disabledStatus?.className).toContain("bg-yellow-100");

      // Check skipped status (yellow)
      const skippedStatus = reportItems[5].querySelector(
        '[data-testid="check-status"]'
      );
      expect(skippedStatus?.className).toContain("bg-yellow-100");

      // Check completed status (blue)
      const completedStatus = reportItems[6].querySelector(
        '[data-testid="check-status"]'
      );
      expect(completedStatus?.className).toContain("bg-blue-100");
    });

    test("should escape HTML in report data", () => {
      const maliciousReports: CheckReport[] = [
        {
          id: '<script>alert("xss")</script>',
          check_slug: '<img src="x" onerror="alert(1)">',
          status: "pass",
          timestamp: "2024-01-15T10:30:00Z",
        },
      ];

      setLatestReports(maliciousReports);

      // Check that HTML tags are escaped in the rendered HTML
      expect(element.innerHTML).not.toContain(
        '<img src="x" onerror="alert(1)">'
      );
      expect(element.innerHTML).toContain("&lt;img");

      // Verify actual displayed text is escaped
      const checkName = element.querySelector('[data-testid="check-name"]');
      expect(checkName?.textContent).toContain(
        '<img src="x" onerror="alert(1)">'
      );
    });
  });
});
