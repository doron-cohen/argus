import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import "./index";
import { flushPromises, waitFor } from "../../../tests/helpers/lit";

type Component = {
  id?: string;
  name: string;
  description?: string;
  owners?: { maintainers?: string[]; team?: string };
};

type CheckReport = {
  id: string;
  check_slug: string;
  status:
    | "pass"
    | "fail"
    | "disabled"
    | "skipped"
    | "unknown"
    | "error"
    | "completed";
  timestamp: string;
};

describe("component-details (unit)", () => {
  let el: any;

  const mockComponent: Component = {
    id: "test-component",
    name: "Test Component",
    description: "This is a test component",
    owners: {
      maintainers: ["john.doe", "jane.smith"],
      team: "Platform Team",
    },
  };

  const mockReports: CheckReport[] = [
    {
      id: "r1",
      check_slug: "unit-tests",
      status: "pass",
      timestamp: "2024-01-15T10:30:00Z",
    },
    {
      id: "r2",
      check_slug: "security-scan",
      status: "fail",
      timestamp: "2024-01-15T10:35:00Z",
    },
    {
      id: "r3",
      check_slug: "code-quality",
      status: "disabled",
      timestamp: "2024-01-15T10:40:00Z",
    },
  ];

  beforeEach(() => {
    el = document.createElement("component-details");
    document.body.appendChild(el);
  });

  afterEach(() => {
    el.remove();
  });

  test("component is defined", () => {
    expect(customElements.get("component-details")).toBeTruthy();
  });

  test("component renders basic structure", async () => {
    el.component = mockComponent;
    el.requestUpdate();
    await flushPromises();

    // Wait a bit more for any async rendering
    await new Promise((resolve) => setTimeout(resolve, 100));

    const componentName = el.shadowRoot?.querySelector(
      '[data-testid="component-name"]',
    );
    expect(componentName).toBeTruthy();
    expect(componentName?.textContent?.trim()).toBe("Test Component");
  });

  describe("component rendering", () => {
    test("renders details when component prop provided", async () => {
      el.component = mockComponent;
      el.requestUpdate();
      await flushPromises();
      await waitFor(
        () => !!el.shadowRoot?.querySelector('[data-testid="component-name"]'),
        500,
      );

      expect(
        el.shadowRoot
          ?.querySelector('[data-testid="component-name"]')
          ?.textContent?.trim(),
      ).toBe("Test Component");
      expect(
        el.shadowRoot
          ?.querySelector('[data-testid="component-id"]')
          ?.textContent?.trim(),
      ).toBe("ID: test-component");
      expect(
        el.shadowRoot
          ?.querySelector('[data-testid="component-description"]')
          ?.textContent?.trim(),
      ).toBe("This is a test component");
      expect(
        el.shadowRoot
          ?.querySelector('[data-testid="component-team"]')
          ?.textContent?.trim(),
      ).toBe("Platform Team");
      expect(
        el.shadowRoot
          ?.querySelector('[data-testid="component-maintainers"]')
          ?.textContent?.trim(),
      ).toBe("john.doe, jane.smith");
    });

    test("uses name as ID when id is missing", async () => {
      const c: Component = { ...mockComponent };
      delete c.id;
      el.component = c;
      el.requestUpdate();
      await flushPromises();
      await waitFor(
        () => !!el.shadowRoot?.querySelector('[data-testid="component-id"]'),
        500,
      );
      expect(
        el.shadowRoot
          ?.querySelector('[data-testid="component-id"]')
          ?.textContent?.trim(),
      ).toBe("ID: Test Component");
    });

    test("renders empty when no component and no error/loading", async () => {
      el.component = null;
      el.requestUpdate();
      await flushPromises();
      expect(
        el.shadowRoot?.querySelector('[data-testid="component-details"]'),
      ).toBeFalsy();
    });

    test("does not render maintainers section when no maintainers", async () => {
      const c: Component = { ...mockComponent };
      delete c.owners?.maintainers;
      el.component = c;
      el.requestUpdate();
      await flushPromises();
      expect(
        el.shadowRoot?.querySelector('[data-testid="component-maintainers"]'),
      ).toBeFalsy();
    });

    test("does not render maintainers section when maintainers array is empty", async () => {
      const c: Component = {
        ...mockComponent,
        owners: { ...mockComponent.owners!, maintainers: [] },
      };
      el.component = c;
      el.requestUpdate();
      await flushPromises();
      expect(
        el.shadowRoot?.querySelector('[data-testid="component-maintainers"]'),
      ).toBeFalsy();
    });
  });

  describe("loading and error", () => {
    test("shows loading skeleton", async () => {
      el.isLoading = true;
      el.requestUpdate();
      await flushPromises();
      await waitFor(
        () =>
          !!el.shadowRoot?.querySelector(
            '[data-testid="component-details-loading"]',
          ),
        500,
      );
      expect(
        el.shadowRoot?.querySelector(
          '[data-testid="component-details-loading"]',
        ),
      ).toBeTruthy();
    });

    test("shows error", async () => {
      el.errorMessage = "Failed to load";
      el.requestUpdate();
      await flushPromises();
      await waitFor(
        () =>
          !!el.shadowRoot?.querySelector(
            '[data-testid="component-details-error"]',
          ),
        500,
      );
      expect(
        el.shadowRoot?.querySelector('[data-testid="component-details-error"]'),
      ).toBeTruthy();
      expect(
        el.shadowRoot
          ?.querySelector('[data-testid="error-title"]')
          ?.textContent?.trim(),
      ).toBe("Error loading component");
    });
  });

  describe("reports", () => {
    beforeEach(async () => {
      el.component = mockComponent;
      el.requestUpdate();
      await flushPromises();
      await waitFor(
        () => !!el.shadowRoot?.querySelector('[data-testid="component-name"]'),
        500,
      );
    });

    test("renders reports list", async () => {
      el.reports = mockReports;
      el.requestUpdate();
      await flushPromises();
      await waitFor(
        () =>
          el.shadowRoot?.querySelectorAll('[data-testid="report-item"]')
            .length > 0,
        500,
      );
      const items = el.shadowRoot?.querySelectorAll(
        '[data-testid="report-item"]',
      );
      expect(items?.length).toBe(3);

      // Check that ui-badge components are used
      const badges = el.shadowRoot?.querySelectorAll("ui-badge");
      expect(badges?.length).toBe(3);

      // Check that badges have correct status attributes
      const firstBadge = badges?.[0] as HTMLElement;
      expect(firstBadge?.getAttribute("status")).toBe("pass");
    });

    test("renders empty reports state", async () => {
      el.reports = [];
      el.requestUpdate();
      await flushPromises();
      await waitFor(
        () => !!el.shadowRoot?.querySelector('[data-testid="no-reports"]'),
        500,
      );
      expect(
        el.shadowRoot?.querySelector('[data-testid="no-reports"]'),
      ).toBeTruthy();
      expect(
        el.shadowRoot?.querySelector('[data-testid="reports-list"]'),
      ).toBeFalsy();
    });

    test("renders loading and error states", async () => {
      el.isReportsLoading = true;
      el.requestUpdate();
      await flushPromises();
      await waitFor(
        () => !!el.shadowRoot?.querySelector('[data-testid="reports-loading"]'),
        500,
      );
      expect(
        el.shadowRoot?.querySelector('[data-testid="reports-loading"]'),
      ).toBeTruthy();

      el.isReportsLoading = false;
      el.reportsErrorMessage = "boom";
      el.requestUpdate();
      await flushPromises();
      await waitFor(
        () => !!el.shadowRoot?.querySelector('[data-testid="reports-error"]'),
        500,
      );
      expect(
        el.shadowRoot?.querySelector('[data-testid="reports-error"]'),
      ).toBeTruthy();
    });

    test("formats timestamps", async () => {
      el.reports = mockReports;
      el.requestUpdate();
      await flushPromises();
      await waitFor(
        () => !!el.shadowRoot?.querySelector('[data-testid="check-timestamp"]'),
        500,
      );
      const ts = el.shadowRoot?.querySelector(
        '[data-testid="check-timestamp"]',
      );
      expect(ts?.textContent).toBeTruthy();
    });
  });
});
