import { expect, fixture, html } from "@open-wc/testing";
import "./index";

describe("sync-status-card", () => {
  const mockStatus = {
    status: "completed",
    lastSync: "2023-12-01T10:00:00Z",
    componentsCount: 5,
    duration: "2.5s",
    lastError: undefined,
  };

  const mockStatusWithError = {
    status: "failed",
    lastSync: "2023-12-01T09:00:00Z",
    componentsCount: 3,
    duration: "1.2s",
    lastError: "Connection timeout",
  };

  it("component is defined", () => {
    expect(customElements.get("sync-status-card")).to.exist;
  });

  it("shows loading state when isLoading is true", async () => {
    const el = await fixture(html`
      <sync-status-card .isLoading=${true}></sync-status-card>
    `);

    const loading = el.shadowRoot?.querySelector("ui-loading-indicator");
    expect(loading).to.exist;
  });

  it("shows error state when error is provided", async () => {
    const el = await fixture(html`
      <sync-status-card .error=${"Connection failed"}></sync-status-card>
    `);

    const errorDiv = el.shadowRoot?.querySelector(".u-text-danger");
    expect(errorDiv).to.exist;
    expect(errorDiv?.textContent).to.include("Error:");
    expect(errorDiv?.textContent).to.include("Connection failed");
  });

  it("shows no status available when status is null", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${null}></sync-status-card>
    `);

    const noStatusDiv = el.shadowRoot?.querySelector(".u-text-muted");
    expect(noStatusDiv).to.exist;
    expect(noStatusDiv?.textContent?.trim()).to.equal("No status available");
  });

  it("renders status badge correctly", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${mockStatus}></sync-status-card>
    `);

    const badge = el.shadowRoot?.querySelector("ui-badge");
    expect(badge).to.exist;
    expect(badge?.textContent?.trim()).to.equal("completed");
  });

  it("renders status badge with correct status mapping", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${mockStatusWithError}></sync-status-card>
    `);

    const badge = el.shadowRoot?.querySelector("ui-badge");
    expect(badge).to.exist;
    expect(badge?.getAttribute("status")).to.equal("fail");
  });

  it("formats timestamp correctly", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${mockStatus}></sync-status-card>
    `);

    // The exact format will depend on locale, but it should contain the date
    const statusText = el.shadowRoot?.textContent;
    expect(statusText).to.include("Last Sync:");
    expect(statusText).to.include("12/1/2023"); // US locale format
  });

  it("displays 'Never' for undefined timestamp", async () => {
    const statusWithoutSync = { ...mockStatus, lastSync: undefined };
    const el = await fixture(html`
      <sync-status-card .status=${statusWithoutSync}></sync-status-card>
    `);

    const statusText = el.shadowRoot?.textContent;
    expect(statusText).to.include("Last Sync:");
    expect(statusText).to.include("Never");
  });

  it("displays components count correctly", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${mockStatus}></sync-status-card>
    `);

    const statusText = el.shadowRoot?.textContent;
    expect(statusText).to.include("Components:");
    expect(statusText).to.include("5");
  });

  it("displays duration correctly", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${mockStatus}></sync-status-card>
    `);

    const statusText = el.shadowRoot?.textContent;
    expect(statusText).to.include("Duration:");
    expect(statusText).to.include("2.5s");
  });

  it("displays 'N/A' for undefined duration", async () => {
    const statusWithoutDuration = { ...mockStatus, duration: undefined };
    const el = await fixture(html`
      <sync-status-card .status=${statusWithoutDuration}></sync-status-card>
    `);

    const statusText = el.shadowRoot?.textContent;
    expect(statusText).to.include("Duration:");
    expect(statusText).to.include("N/A");
  });

  it("shows last error when present", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${mockStatusWithError}></sync-status-card>
    `);

    const statusText = el.shadowRoot?.textContent;
    expect(statusText).to.include("Last Error:");
    expect(statusText).to.include("Connection timeout");
  });

  it("does not show last error section when no error", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${mockStatus}></sync-status-card>
    `);

    const statusText = el.shadowRoot?.textContent;
    expect(statusText).to.not.include("Last Error:");
  });

  it("maps status values correctly", async () => {
    const testCases = [
      { status: "idle", expectedBadgeStatus: "default" },
      { status: "running", expectedBadgeStatus: "default" },
      { status: "completed", expectedBadgeStatus: "pass" },
      { status: "failed", expectedBadgeStatus: "fail" },
      { status: "unknown", expectedBadgeStatus: "default" },
    ];

    for (const testCase of testCases) {
      const status = { ...mockStatus, status: testCase.status };
      const el = await fixture(html`
        <sync-status-card .status=${status}></sync-status-card>
      `);

      const badge = el.shadowRoot?.querySelector("ui-badge");
      expect(badge?.getAttribute("status")).to.equal(
        testCase.expectedBadgeStatus,
      );
    }
  });

  it("shows loading indicator when isLoading is true even with null status", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${null} .isLoading=${true}></sync-status-card>
    `);

    const loading = el.shadowRoot?.querySelector("ui-loading-indicator");
    expect(loading).to.exist;
    expect(loading?.textContent?.trim()).to.equal("Loading...");
  });

  it("shows 'No status available' when status is null and not loading", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${null} .isLoading=${false}></sync-status-card>
    `);

    const message = el.shadowRoot?.textContent?.trim();
    expect(message).to.equal("No status available");
  });

  it("shows 'Status data loading...' for empty status object", async () => {
    const emptyStatus = {};
    const el = await fixture(html`
      <sync-status-card
        .status=${emptyStatus}
        .isLoading=${false}
      ></sync-status-card>
    `);

    const message = el.shadowRoot?.textContent?.trim();
    expect(message).to.equal("Status data loading...");
  });

  it("shows status data when valid status is provided", async () => {
    const el = await fixture(html`
      <sync-status-card .status=${mockStatus}></sync-status-card>
    `);

    const badge = el.shadowRoot?.querySelector("ui-badge");
    expect(badge).to.exist;
    expect(badge?.textContent?.trim()).to.equal("completed");

    const statusText = el.shadowRoot?.textContent;
    expect(statusText).to.include("Last Sync:");
    expect(statusText).to.include("Components:");
    expect(statusText).to.include("Duration:");
  });
});
