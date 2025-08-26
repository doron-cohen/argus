import { LitElement, html, nothing } from "lit";
import { customElement, state } from "lit/decorators.js";
import { loadSyncSources } from "./data";
import { resetSettings } from "./store";
import "../../ui/primitives/page-container.js";
import "../../ui/components/ui-page-header.js";
import "../../ui/components/ui-card.js";
import "../../ui/components/ui-empty-state.js";
import "../../ui/components/ui-loading-indicator.js";
import "../../ui/components/ui-description-list.js";
import "../../components/sync-status-card/index.js";
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
      }),
    );

    this.unsubscribers.push(
      sourceStatuses.subscribe((value) => {
        this.statuses = value;
        this.requestUpdate();
      }),
    );

    this.unsubscribers.push(
      settingsLoading.subscribe((value) => {
        this.isLoading = value;
        this.requestUpdate();
      }),
    );

    this.unsubscribers.push(
      settingsError.subscribe((value) => {
        this.error = value;
        this.requestUpdate();
      }),
    );

    this.unsubscribers.push(
      statusesLoading.subscribe((value) => {
        this.statusLoading = value;
        this.requestUpdate();
      }),
    );

    this.unsubscribers.push(
      statusesError.subscribe((value) => {
        this.statusErrors = value;
        this.requestUpdate();
      }),
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
              padding="md"
              data-testid="sync-source-${source.id || "unknown"}"
            >
              <div slot="header">
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

              <div>
                <div
                  class="grid grid-cols-1 lg:grid-cols-2 u-gap-6"
                  style="max-width: 100%"
                >
                  <div>
                    <h4 class="u-section-title">Configuration</h4>
                    ${this.renderSourceConfig(source)}
                  </div>
                  <div>
                    <h4 class="u-section-title">Status</h4>
                    <sync-status-card
                      .status=${this.statuses[source.id!] || null}
                      .isLoading=${this.statusLoading[source.id!] || false}
                      .error=${this.statusErrors[source.id!] || null}
                    ></sync-status-card>
                  </div>
                </div>
              </div>
            </ui-card>
          `,
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
