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
  let el: HTMLElement;

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

  describe("component rendering", () => {
    test("renders details when component prop provided", async () => {
      (el as any).component = mockComponent;
      (el as any).requestUpdate?.();
      await flushPromises();
      await waitFor(
        () => !!el.querySelector('[data-testid="component-name"]'),
        500,
      );

      expect(
        el.querySelector('[data-testid="component-name"]')?.textContent?.trim(),
      ).toBe("Test Component");
      expect(
        el.querySelector('[data-testid="component-id"]')?.textContent?.trim(),
      ).toBe("ID: test-component");
      expect(
        el
          .querySelector('[data-testid="component-description"]')
          ?.textContent?.trim(),
      ).toBe("This is a test component");
      expect(
        el.querySelector('[data-testid="component-team"]')?.textContent?.trim(),
      ).toBe("Platform Team");
      expect(
        el
          .querySelector('[data-testid="component-maintainers"]')
          ?.textContent?.trim(),
      ).toBe("john.doe, jane.smith");
    });

    test("uses name as ID when id is missing", async () => {
      const c: Component = { ...mockComponent };
      delete c.id;
      (el as any).component = c;
      (el as any).requestUpdate?.();
      await flushPromises();
      expect(
        el.querySelector('[data-testid="component-id"]')?.textContent?.trim(),
      ).toBe("ID: Test Component");
    });

    test("renders empty when no component and no error/loading", async () => {
      (el as any).component = null;
      (el as any).requestUpdate?.();
      await flushPromises();
      expect(el.querySelector('[data-testid="component-details"]')).toBeFalsy();
    });

    test("includes back link", async () => {
      (el as any).component = mockComponent;
      (el as any).requestUpdate?.();
      await flushPromises();
      const back = el.querySelector(
        '[data-testid="back-to-components"]',
      ) as HTMLAnchorElement | null;
      expect(back).toBeTruthy();
      expect(back?.getAttribute("href")).toBe("/");
    });
  });

  describe("loading and error", () => {
    test("shows loading skeleton", async () => {
      (el as any).isLoading = true;
      (el as any).requestUpdate?.();
      await flushPromises();
      expect(
        el.querySelector('[data-testid="component-details-loading"]'),
      ).toBeTruthy();
    });

    test("shows error", async () => {
      (el as any).errorMessage = "Failed to load";
      (el as any).requestUpdate?.();
      await flushPromises();
      expect(
        el.querySelector('[data-testid="component-details-error"]'),
      ).toBeTruthy();
      expect(
        el.querySelector('[data-testid="error-title"]')?.textContent?.trim(),
      ).toBe("Error loading component");
    });
  });

  describe("reports", () => {
    beforeEach(async () => {
      (el as any).component = mockComponent;
      (el as any).requestUpdate?.();
      await flushPromises();
    });

    test("renders reports list", async () => {
      (el as any).reports = mockReports;
      (el as any).requestUpdate?.();
      await flushPromises();
      const items = el.querySelectorAll('[data-testid="report-item"]');
      expect(items.length).toBe(3);
      // check badge class coloring via host classes
      const firstStatus = items[0].querySelector(
        '[data-testid="check-status"]',
      ) as HTMLElement | null;
      expect(firstStatus?.className).toContain("bg-green-100");
    });

    test("renders empty reports state", async () => {
      (el as any).reports = [];
      (el as any).requestUpdate?.();
      await flushPromises();
      expect(el.querySelector('[data-testid="no-reports"]')).toBeTruthy();
      expect(el.querySelector('[data-testid="reports-list"]')).toBeFalsy();
    });

    test("renders loading and error states", async () => {
      (el as any).isReportsLoading = true;
      (el as any).requestUpdate?.();
      await flushPromises();
      expect(el.querySelector('[data-testid="reports-loading"]')).toBeTruthy();

      (el as any).isReportsLoading = false;
      (el as any).reportsErrorMessage = "boom";
      (el as any).requestUpdate?.();
      await flushPromises();
      expect(el.querySelector('[data-testid="reports-error"]')).toBeTruthy();
    });

    test("formats timestamps", async () => {
      (el as any).reports = mockReports;
      (el as any).requestUpdate?.();
      await flushPromises();
      const ts = el.querySelector('[data-testid="check-timestamp"]');
      expect(ts?.textContent).toBeTruthy();
    });
  });
});
