import { LitElement, html, nothing } from "lit";
import { customElement, state } from "lit/decorators.js";
import { loadSyncSources } from "./data";
import { resetSettings } from "./store";
import "../ui/primitives/page-container.js";
import "../ui/components/ui-page-header.js";
import "../ui/components/ui-card.js";
import "../ui/components/ui-badge.js";
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
      return html`
        <div class="space-y-2">
          <div>
            <span class="font-medium text-gray-700">Repository:</span>
            <span class="text-gray-900">${config.url || "N/A"}</span>
          </div>
          <div>
            <span class="font-medium text-gray-700">Branch:</span>
            <span class="text-gray-900">${config.branch || "N/A"}</span>
          </div>
          ${config.basePath
            ? html`
                <div>
                  <span class="font-medium text-gray-700">Base Path:</span>
                  <span class="text-gray-900">${config.basePath}</span>
                </div>
              `
            : nothing}
        </div>
      `;
    } else if (source.config) {
      const config = source.config as FilesystemSourceConfig;
      return html`
        <div class="space-y-2">
          <div>
            <span class="font-medium text-gray-700">Path:</span>
            <span class="text-gray-900">${config.path}</span>
          </div>
          ${config.basePath
            ? html`
                <div>
                  <span class="font-medium text-gray-700">Base Path:</span>
                  <span class="text-gray-900">${config.basePath}</span>
                </div>
              `
            : nothing}
        </div>
      `;
    }
  }

  private renderSourceStatus(source: SyncSource) {
    const sourceId = source.id;
    if (sourceId === undefined) {
      return html`
        <div class="text-gray-500 text-sm">No source ID available</div>
      `;
    }

    const status = this.statuses[sourceId];
    const isLoading = this.statusLoading[sourceId];
    const error = this.statusErrors[sourceId];

    if (isLoading) {
      return html`
        <div class="flex items-center space-x-2">
          <div
            class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"
          ></div>
          <span class="text-gray-500">Loading...</span>
        </div>
      `;
    }

    if (error) {
      return html`
        <div class="text-red-600 text-sm">
          <span class="font-medium">Error:</span> ${error}
        </div>
      `;
    }

    if (!status) {
      return html`
        <div class="text-gray-500 text-sm">No status available</div>
      `;
    }

    return html`
      <div class="space-y-3">
        <div class="flex items-center space-x-2">
          <ui-badge
            status=${this.getStatusBadgeStatus(status.status || "unknown")}
          >
            ${status.status || "unknown"}
          </ui-badge>
        </div>

        <div class="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span class="font-medium text-gray-700">Last Sync:</span>
            <div class="text-gray-900">
              ${this.formatTimestamp(status.lastSync || undefined)}
            </div>
          </div>
          <div>
            <span class="font-medium text-gray-700">Components:</span>
            <div class="text-gray-900">${status.componentsCount}</div>
          </div>
          <div>
            <span class="font-medium text-gray-700">Duration:</span>
            <div class="text-gray-900">
              ${this.formatDuration(status.duration || undefined)}
            </div>
          </div>
        </div>

        ${status.lastError
          ? html`
              <div class="mt-2">
                <span class="font-medium text-red-700">Last Error:</span>
                <div class="text-red-600 text-sm mt-1">${status.lastError}</div>
              </div>
            `
          : nothing}
      </div>
    `;
  }

  private renderSources() {
    if (this.sources.length === 0) {
      return html`
        <div class="text-center py-8">
          <div class="text-gray-500 text-lg">No sync sources configured</div>
          <div class="text-gray-400 text-sm mt-2">
            Configure sync sources to see them here
          </div>
        </div>
      `;
    }

    return html`
      <div class="space-y-6">
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
                    <h3 class="text-lg font-medium text-gray-900">
                      ${source.type === "git" ? "Git Repository" : "Filesystem"}
                      #${source.id || "unknown"}
                    </h3>
                    <p class="text-sm text-gray-500">
                      Sync interval: ${source.interval}
                    </p>
                  </div>
                </div>
              </div>

              <div class="px-6 py-4">
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div>
                    <h4 class="font-medium text-gray-900 mb-3">
                      Configuration
                    </h4>
                    ${this.renderSourceConfig(source)}
                  </div>
                  <div>
                    <h4 class="font-medium text-gray-900 mb-3">Status</h4>
                    ${this.renderSourceStatus(source)}
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
              <div class="flex justify-center items-center py-8">
                <div class="text-gray-500">Loading settings...</div>
              </div>
            `
          : this.error
            ? html`
                <ui-card variant="outlined" padding="md">
                  <div class="text-red-600">
                    <h2 class="text-red-800 font-semibold mb-2">
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
