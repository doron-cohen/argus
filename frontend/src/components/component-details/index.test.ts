import { expect, fixture, html } from "@open-wc/testing";
import "./index";

// Import subcomponents for testing
import "./metadata.js";
import "./reports.js";

// Import UI components used by component-details
import "../../ui/components/ui-spinner.js";
import "../../ui/components/ui-alert.js";

import type { ComponentDetails } from "./index";

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
    const el = await fixture<ComponentDetails>(html`
      <component-details .errorMessage=${"Test error"}></component-details>
    `);
    await el.updateComplete;

    const error = el.shadowRoot?.querySelector(
      '[data-testid="component-details-error"]',
    );
    expect(error).to.exist;

    // Check that the ui-alert component contains the error message
    const uiAlert = error?.querySelector("ui-alert");
    expect(uiAlert).to.exist;

    // Check that the error message is passed to the ui-alert component
    expect((uiAlert as any)?.message).to.equal("Test error");
  });

  it("renders empty state when no component is provided", async () => {
    const el = await fixture(html` <component-details></component-details> `);

    expect(el.shadowRoot?.textContent).to.include(
      "No component data available",
    );
  });

  it("renders component details with subcomponents when component is provided", async () => {
    const el = await fixture<ComponentDetails>(html`
      <component-details .component=${mockComponent}></component-details>
    `);
    await el.updateComplete;

    // Check that subcomponents are rendered
    const metadataComponent =
      el.shadowRoot?.querySelector("component-metadata");
    expect(metadataComponent).to.exist;

    const reportsComponent = el.shadowRoot?.querySelector("component-reports");
    expect(reportsComponent).to.exist;

    // Check that component data is passed to metadata subcomponent
    expect(metadataComponent?.component).to.equal(mockComponent);
  });

  it("renders reports subcomponent with proper props when reports are provided", async () => {
    const el = await fixture<ComponentDetails>(html`
      <component-details
        .component=${mockComponent}
        .reports=${mockReports}
        .isReportsLoading=${false}
        .reportsErrorMessage=${null}
      ></component-details>
    `);
    await el.updateComplete;

    const reportsComponent = el.shadowRoot?.querySelector("component-reports");
    expect(reportsComponent).to.exist;

    // Check that reports data is passed to reports subcomponent
    expect(reportsComponent?.reports).to.equal(mockReports);
    expect(reportsComponent?.isLoading).to.equal(false);
    expect(reportsComponent?.errorMessage).to.equal(null);
  });

  it("passes reports loading state to reports subcomponent", async () => {
    const el = await fixture<ComponentDetails>(html`
      <component-details
        .component=${mockComponent}
        .isReportsLoading=${true}
      ></component-details>
    `);
    await el.updateComplete;

    const reportsComponent = el.shadowRoot?.querySelector("component-reports");
    expect(reportsComponent).to.exist;
    expect(reportsComponent?.isLoading).to.equal(true);
  });

  it("passes reports error message to reports subcomponent", async () => {
    const errorMsg = "Failed to load reports";
    const el = await fixture<ComponentDetails>(html`
      <component-details
        .component=${mockComponent}
        .reportsErrorMessage=${errorMsg}
      ></component-details>
    `);
    await el.updateComplete;

    const reportsComponent = el.shadowRoot?.querySelector("component-reports");
    expect(reportsComponent).to.exist;
    expect(reportsComponent?.errorMessage).to.equal(errorMsg);
  });
});
