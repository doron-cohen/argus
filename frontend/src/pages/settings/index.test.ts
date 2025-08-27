import { expect, fixture, html } from "@open-wc/testing";
import { SettingsPage } from "./index";
import {
  syncSources,
  sourceStatuses,
  settingsLoading,
  settingsError,
} from "./store";

// Import UI components needed for the page
import "../../ui/primitives/page-container.js";
import "../../ui/components/ui-page-header.js";
import "../../ui/components/ui-card.js";
import "../../ui/components/ui-card-header.js";
import "../../ui/components/ui-info-row.js";
import "../../ui/components/ui-badge.js";
import "../../ui/components/ui-empty-state.js";
import "../../ui/components/ui-loading-indicator.js";
import "../../components/sync-status-card/index.js";

describe("SettingsPage", () => {
  let element: SettingsPage;

  beforeEach(async () => {
    // Reset stores before each test
    syncSources.set([]);
    sourceStatuses.set({});
    settingsLoading.set(false);
    settingsError.set(null);

    element = await fixture(html`<settings-page></settings-page>`);
    await element.updateComplete;
  });

  it("can be created", async () => {
    expect(element).to.exist;
    expect(element.tagName.toLowerCase()).to.equal("settings-page");
  });

  it("renders with page title and description", () => {
    const title = element.shadowRoot?.querySelector(
      '[data-testid="page-title"]',
    );
    const description = element.shadowRoot?.querySelector(
      '[data-testid="page-description"]',
    );

    expect(title).to.exist;
    expect(title?.textContent?.trim()).to.equal("Settings");
    expect(description).to.exist;
    expect(description?.textContent?.trim()).to.equal(
      "Sync source configuration and status information",
    );
  });

  it("shows loading state when loading", async () => {
    settingsLoading.set(true);
    await element.updateComplete;
    await element.updateComplete; // Wait for store changes to propagate

    const loadingIndicator = element.shadowRoot?.querySelector(
      "ui-loading-indicator",
    );
    expect(loadingIndicator).to.exist;
  });

  it("shows error state when there's an error", async () => {
    settingsError.set("Failed to load settings");
    await element.updateComplete;
    await element.updateComplete; // Wait for store changes to propagate

    const errorCard = element.shadowRoot?.querySelector("ui-card");
    expect(errorCard).to.exist;
    const errorText = errorCard?.textContent;
    expect(errorText).to.include("Failed to load settings");
  });

  it("shows empty state when no sources configured", async () => {
    syncSources.set([]);
    await element.updateComplete;
    await element.updateComplete; // Wait for store changes to propagate

    const emptyState = element.shadowRoot?.querySelector("ui-empty-state");
    expect(emptyState).to.exist;
  });

  it("renders git source configuration correctly", async () => {
    const gitSource = {
      id: 1,
      type: "git" as const,
      config: {
        url: "https://github.com/example/repo",
        branch: "main",
        basePath: "/components",
      },
      interval: "5m",
    };

    syncSources.set([gitSource]);
    await element.updateComplete;
    await element.updateComplete; // Wait for store changes to propagate

    const sourceElement = element.shadowRoot?.querySelector(
      '[data-testid="sync-source-1"]',
    );
    expect(sourceElement).to.exist;

    const sourceTitle = sourceElement?.querySelector("h3");
    expect(sourceTitle?.textContent?.trim()).to.include("Git Repository #1");

    const intervalText = sourceElement?.querySelector(".u-text-muted");
    expect(intervalText?.textContent?.trim()).to.include("Sync interval: 5m");

    const urlText = sourceElement?.textContent;
    expect(urlText).to.include("https://github.com/example/repo");
    expect(urlText).to.include("main");
    expect(urlText).to.include("/components");
  });

  it("renders filesystem source configuration correctly", async () => {
    const fsSource = {
      id: 2,
      type: "filesystem" as const,
      config: {
        path: "/var/components",
        basePath: "/services",
      },
      interval: "1h",
    };

    syncSources.set([fsSource]);
    await element.updateComplete;
    await element.updateComplete; // Wait for store changes to propagate

    const sourceElement = element.shadowRoot?.querySelector(
      '[data-testid="sync-source-2"]',
    );
    expect(sourceElement).to.exist;

    const sourceTitle = sourceElement?.querySelector("h3");
    expect(sourceTitle?.textContent?.trim()).to.include("Filesystem #2");

    const pathText = sourceElement?.textContent;
    expect(pathText).to.include("/var/components");
    expect(pathText).to.include("/services");
  });

  it("renders source status correctly", async () => {
    const source = {
      id: 1,
      type: "git" as const,
      config: { url: "https://example.com", branch: "main" },
      interval: "5m",
    };

    const status = {
      sourceId: 1,
      status: "completed" as const,
      lastSync: "2023-01-01T12:00:00Z",
      componentsCount: 5,
      duration: "30s",
    };

    syncSources.set([source]);
    sourceStatuses.set({ 1: status });
    await element.updateComplete;
    await element.updateComplete; // Wait for store changes to propagate

    const sourceElement = element.shadowRoot?.querySelector(
      '[data-testid="sync-source-1"]',
    );
    const statusText = sourceElement?.textContent;

    expect(statusText).to.include("completed");
    expect(statusText).to.include("5"); // components count
    expect(statusText).to.include("30s"); // duration
  });

  it("handles missing status gracefully", async () => {
    const source = {
      id: 1,
      type: "git" as const,
      config: { url: "https://example.com", branch: "main" },
      interval: "5m",
    };

    syncSources.set([source]);
    sourceStatuses.set({});
    await element.updateComplete;
    await element.updateComplete; // Wait for store changes to propagate

    const sourceElement = element.shadowRoot?.querySelector(
      '[data-testid="sync-source-1"]',
    );
    const statusText = sourceElement?.textContent;

    expect(statusText).to.include("No status available");
  });
});
