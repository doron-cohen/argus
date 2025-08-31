import { expect, fixture, html } from "@open-wc/testing";
import "./reports";

// Import UI components used by component-reports
import "../../ui/components/ui-card.js";
import "../../ui/components/ui-badge.js";
import "../../ui/components/ui-spinner.js";
import "../../ui/components/ui-alert.js";

import type { CheckReport } from "../../api/services/components/client";
import type { ComponentReports } from "./reports";

describe("component-reports", () => {
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
    expect(customElements.get("component-reports")).to.exist;
  });

  it("renders loading state when isLoading is true", async () => {
    const el = await fixture<ComponentReports>(html`
      <component-reports .isLoading=${true}></component-reports>
    `);
    await el.updateComplete;

    const loading = el.shadowRoot?.querySelector(
      '[data-testid="reports-loading"]',
    );
    expect(loading).to.exist;
    expect(loading?.textContent).to.include("Loading reports");
  });

  it("renders error state when errorMessage is provided", async () => {
    const el = await fixture<ComponentReports>(html`
      <component-reports .errorMessage=${"Test error"}></component-reports>
    `);
    await el.updateComplete;

    const error = el.shadowRoot?.querySelector('[data-testid="reports-error"]');
    expect(error).to.exist;

    // Check that the ui-alert component contains the error message
    const uiAlert = error?.querySelector("ui-alert");
    expect(uiAlert).to.exist;

    // Check that the error message is passed to the ui-alert component
    expect((uiAlert as any)?.message).to.equal("Test error");
  });

  it("renders empty state when no reports are provided", async () => {
    const el = await fixture(html` <component-reports></component-reports> `);

    expect(el.shadowRoot?.textContent).to.include(
      "No reports available for this component",
    );
  });

  it("renders empty state when reports array is empty", async () => {
    const el = await fixture(html`
      <component-reports .reports=${[]}></component-reports>
    `);

    expect(el.shadowRoot?.textContent).to.include(
      "No reports available for this component",
    );
  });

  it("renders reports list when reports are provided", async () => {
    const el = await fixture<ComponentReports>(html`
      <component-reports .reports=${mockReports}></component-reports>
    `);
    await el.updateComplete;

    const reports = el.shadowRoot?.querySelectorAll(
      '[data-testid="report-item"]',
    );
    expect(reports).to.have.length(3);

    const reportsLabel = el.shadowRoot?.querySelector(
      '[data-testid="reports-label"]',
    );
    expect(reportsLabel).to.exist;
    expect(reportsLabel?.textContent?.trim()).to.equal("Latest Quality Checks");
  });

  it("renders correct status badges for reports", async () => {
    const el = await fixture<ComponentReports>(html`
      <component-reports .reports=${mockReports}></component-reports>
    `);
    await el.updateComplete;

    const passBadge = el.shadowRoot?.querySelector('ui-badge[status="pass"]');
    const failBadge = el.shadowRoot?.querySelector('ui-badge[status="fail"]');
    const disabledBadge = el.shadowRoot?.querySelector(
      'ui-badge[status="disabled"]',
    );

    expect(passBadge).to.exist;
    expect(failBadge).to.exist;
    expect(disabledBadge).to.exist;
  });

  it("renders check names correctly", async () => {
    const el = await fixture<ComponentReports>(html`
      <component-reports .reports=${mockReports}></component-reports>
    `);
    await el.updateComplete;

    const checkNames = el.shadowRoot?.querySelectorAll(
      '[data-testid="check-name"]',
    );
    expect(checkNames).to.have.length(3);

    expect(checkNames?.[0]?.textContent?.trim()).to.equal("unit-tests");
    expect(checkNames?.[1]?.textContent?.trim()).to.equal("security-scan");
    expect(checkNames?.[2]?.textContent?.trim()).to.equal("code-quality");
  });

  it("renders timestamps correctly", async () => {
    const el = await fixture<ComponentReports>(html`
      <component-reports .reports=${mockReports}></component-reports>
    `);
    await el.updateComplete;

    const timestamps = el.shadowRoot?.querySelectorAll(
      '[data-testid="check-timestamp"]',
    );
    expect(timestamps).to.have.length(3);

    // Check that timestamps are formatted (they should be converted from ISO to locale string)
    timestamps?.forEach((timestamp, index) => {
      expect(timestamp?.textContent?.trim()).to.not.equal("");
      expect(timestamp?.textContent?.trim()).to.not.equal(
        mockReports[index].timestamp,
      ); // Should be different from raw ISO
    });
  });
});
