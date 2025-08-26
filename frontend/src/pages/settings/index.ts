import { LitElement, html, nothing } from "lit";
import { customElement, state } from "lit/decorators.js";
import { loadSyncSources } from "./data";
import { resetSettings } from "./store";
import "../../ui/primitives/page-container.js";
import "../../ui/components/ui-page-header.js";
import "../../ui/components/ui-card.js";
import "../../ui/components/ui-badge.js";
import "../../ui/components/ui-empty-state.js";
import "../../ui/components/ui-loading-indicator.js";
import "../../ui/components/ui-description-list.js";
import {
  syncSources,
  sourceStatuses,
  settingsLoading,
  statusesLoading,
  settingsError,
  statusesError,
  type SyncSource,
  type SyncStatus,
} from "./store";
import type {
  GitSourceConfig,
  FilesystemSourceConfig,
} from "../../api/services/sync/client";
import type { DescriptionItem } from "../../ui/components/ui-description-list";

@customElement("settings-page")
export class SettingsPage extends LitElement {
  @state()
  sources: readonly SyncSource[] = [];

  @state()
  statuses: Record<number, SyncStatus> = {};

  @state()
  isLoading = false;

  @state()
  error: string | null = null;

  @state()
  statusLoading: Record<number, boolean> = {};

  @state()
  statusErrors: Record<number, string | null> = {};

  private unsubscribers: Array<() => void> = [];

  async connectedCallback(): Promise<void> {
    super.connectedCallback();

    // Subscribe to stores
    this.unsubscribers.push(
      syncSources.subscribe((value) => {
        this.sources = value;
        this.requestUpdate();
      })
    );

    this.unsubscribers.push(
      sourceStatuses.subscribe((value) => {
        this.statuses = value;
        this.requestUpdate();
      })
    );

    this.unsubscribers.push(
      settingsLoading.subscribe((value) => {
        this.isLoading = value;
        this.requestUpdate();
      })
    );

    this.unsubscribers.push(
      settingsError.subscribe((value) => {
        this.error = value;
        this.requestUpdate();
      })
    );

    this.unsubscribers.push(
      statusesLoading.subscribe((value) => {
        this.statusLoading = value;
        this.requestUpdate();
      })
    );

    this.unsubscribers.push(
      statusesError.subscribe((value) => {
        this.statusErrors = value;
        this.requestUpdate();
      })
    );

    await this.load();
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.unsubscribers.forEach((unsub) => unsub());
    resetSettings();
  }

  private async load(): Promise<void> {
    await loadSyncSources();
  }

  formatTimestamp(timestamp?: string): string {
    if (!timestamp) return "Never";
    return new Date(timestamp).toLocaleString();
  }

  formatDuration(duration?: string): string {
    if (!duration) return "N/A";
    return duration;
  }

  getStatusBadgeStatus(status: string): string {
    switch (status) {
      case "idle":
        return "default";
      case "running":
        return "default";
      case "completed":
        return "pass";
      case "failed":
        return "fail";
      default:
        return "default";
    }
  }

  private renderSourceConfig(source: SyncSource) {
    if (source.type === "git" && source.config) {
      const config = source.config as GitSourceConfig;
      const items = [
        { label: "Repository", value: config.url || "N/A" },
        { label: "Branch", value: config.branch || "N/A" },
        ...(config.basePath
          ? [{ label: "Base Path", value: config.basePath }]
          : []),
      ];
      return html`<ui-description-list .items=${items}></ui-description-list>`;
    } else if (source.config) {
      const config = source.config as FilesystemSourceConfig;
      const items = [
        { label: "Path", value: config.path },
        ...(config.basePath
          ? [{ label: "Base Path", value: config.basePath }]
          : []),
      ];
      return html`<ui-description-list .items=${items}></ui-description-list>`;
    }
    return html``;
  }

  private renderSourceStatus(source: SyncSource) {
    const sourceId = source.id;
    if (sourceId === undefined) {
      return html`
        <div class="u-text-muted text-sm">No source ID available</div>
      `;
    }

    const status = this.statuses[sourceId];
    const isLoading = this.statusLoading[sourceId];
    const error = this.statusErrors[sourceId];

    if (isLoading) {
      return html`
        <ui-loading-indicator
          message="Loading..."
          size="sm"
        ></ui-loading-indicator>
      `;
    }

    if (error) {
      return html`
        <div class="u-text-danger text-sm">
          <span class="u-font-medium">Error:</span> ${error}
        </div>
      `;
    }

    if (!status) {
      return html`
        <div class="u-text-muted text-sm">No status available</div>
      `;
    }

    return html`
      <div class="u-stack-3">
        <div class="flex items-center space-x-2">
          <ui-badge
            status=${this.getStatusBadgeStatus(status.status || "unknown")}
          >
            ${status.status || "unknown"}
          </ui-badge>
        </div>

        <div class="u-grid-2 text-sm">
          <div>
            <span class="u-font-medium u-text-secondary">Last Sync:</span>
            <div class="u-text-primary">
              ${this.formatTimestamp(status.lastSync || undefined)}
            </div>
          </div>
          <div>
            <span class="u-font-medium u-text-secondary">Components:</span>
            <div class="u-text-primary">${status.componentsCount}</div>
          </div>
          <div>
            <span class="u-font-medium u-text-secondary">Duration:</span>
            <div class="u-text-primary">
              ${this.formatDuration(status.duration || undefined)}
            </div>
          </div>
        </div>

        ${status.lastError
          ? html`
              <div class="mt-2">
                <span class="u-font-medium u-text-secondary">Last Error:</span>
                <div class="u-text-danger text-sm mt-1">
                  ${status.lastError}
                </div>
              </div>
            `
          : nothing}
      </div>
    `;
  }

  private renderSources() {
    if (this.sources.length === 0) {
      return html`
        <ui-empty-state
          title="No sync sources configured"
          description="Configure sync sources to see them here"
        ></ui-empty-state>
      `;
    }

    return html`
      <div class="u-stack-6">
        ${this.sources.map(
          (source) => html`
            <ui-card
              variant="default"
              padding="none"
              data-testid="sync-source-${source.id || "unknown"}"
            >
              <div slot="header" class="px-6 py-4">
                <div class="flex items-center justify-between">
                  <div>
                    <h3 class="text-lg u-font-medium u-text-primary">
                      ${source.type === "git" ? "Git Repository" : "Filesystem"}
                      #${source.id || "unknown"}
                    </h3>
                    <p class="text-sm u-text-muted">
                      Sync interval: ${source.interval}
                    </p>
                  </div>
                </div>
              </div>

              <div class="px-6 py-4">
                <div class="grid grid-cols-1 lg:grid-cols-2 u-gap-6">
                  <div>
                    <h4 class="u-section-title">Configuration</h4>
                    ${this.renderSourceConfig(source)}
                  </div>
                  <div>
                    <h4 class="u-section-title">Status</h4>
                    ${this.renderSourceStatus(source)}
                  </div>
                </div>
              </div>
            </ui-card>
          `
        )}
      </div>
    `;
  }

  render() {
    return html`
      <ui-page-container max-width="xl" padding="lg">
        <ui-page-header
          title="Settings"
          description="Sync source configuration and status information"
          size="lg"
        ></ui-page-header>

        ${this.isLoading
          ? html`
              <ui-loading-indicator
                message="Loading settings..."
              ></ui-loading-indicator>
            `
          : this.error
            ? html`
                <ui-card variant="outlined" padding="md">
                  <div class="u-text-danger">
                    <h2 class="u-text-danger u-font-semibold mb-2">
                      Error loading settings
                    </h2>
                    <p>${this.error}</p>
                  </div>
                </ui-card>
              `
            : this.renderSources()}
      </ui-page-container>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "settings-page": SettingsPage;
  }
}

// Only define if not already defined (handles hot reload scenarios)
if (!customElements.get("settings-page")) {
  customElements.define("settings-page", SettingsPage);
}
