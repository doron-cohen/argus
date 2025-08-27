import { expect, fixture, html } from "@open-wc/testing";
import "./index";

// Import UI components used by component-details
import "../../ui/components/ui-badge.js";
import "../../ui/components/ui-card.js";
import "../../ui/components/ui-card-header.js";
import "../../ui/components/ui-info-row.js";
import "../../ui/components/ui-spinner.js";
import "../../ui/components/ui-alert.js";

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

describe("component-details", () => {
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

  it("component is defined", () => {
    expect(customElements.get("component-details")).to.exist;
  });

  it("renders loading state when isLoading is true", async () => {
    const el = await fixture(html`
      <component-details .isLoading=${true}></component-details>
    `);

    const loading = el.shadowRoot?.querySelector(
      '[data-testid="component-details-loading"]',
    );
    expect(loading).to.exist;
    expect(loading?.textContent).to.include("Loading component details");
  });

  it("renders error state when errorMessage is provided", async () => {
    const el = await fixture(html`
      <component-details .errorMessage=${"Test error"}></component-details>
    `);

    const error = el.shadowRoot?.querySelector(
      '[data-testid="component-details-error"]',
    );
    expect(error).to.exist;
    expect(error?.textContent).to.include("Test error");
  });

  it("renders empty state when no component is provided", async () => {
    const el = await fixture(html` <component-details></component-details> `);

    expect(el.shadowRoot?.textContent).to.include(
      "No component data available",
    );
  });

  it("renders component details when component is provided", async () => {
    const el = await fixture(html`
      <component-details .component=${mockComponent}></component-details>
    `);

    const name = el.shadowRoot?.querySelector('[data-testid="component-name"]');
    expect(name).to.exist;
    expect(name?.textContent?.trim()).to.equal("Test Component");

    const description = el.shadowRoot?.querySelector(
      '[data-testid="component-description"]',
    );
    expect(description).to.exist;
    expect(description?.textContent?.trim()).to.equal(
      "This is a test component",
    );
  });

  it("renders reports when provided", async () => {
    const el = await fixture(html`
      <component-details
        .component=${mockComponent}
        .reports=${mockReports}
      ></component-details>
    `);

    const reports = el.shadowRoot?.querySelectorAll(
      '[data-testid="report-item"]',
    );
    expect(reports).to.have.length(3);
  });

  it("renders correct status badges for reports", async () => {
    const el = await fixture(html`
      <component-details
        .component=${mockComponent}
        .reports=${mockReports}
      ></component-details>
    `);

    const passBadge = el.shadowRoot?.querySelector('ui-badge[status="pass"]');
    const failBadge = el.shadowRoot?.querySelector('ui-badge[status="fail"]');
    const disabledBadge = el.shadowRoot?.querySelector(
      'ui-badge[status="disabled"]',
    );

    expect(passBadge).to.exist;
    expect(failBadge).to.exist;
    expect(disabledBadge).to.exist;
  });
});
