import { LitElement, html, nothing } from "lit";
import { customElement, state } from "lit/decorators.js";
import { loadSyncSources } from "./data";
import { resetSettings } from "./store";
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

  getStatusBadgeClass(status: string): string {
    switch (status) {
      case "idle":
        return "bg-gray-100 text-gray-800";
      case "running":
        return "bg-blue-100 text-blue-800";
      case "completed":
        return "bg-green-100 text-green-800";
      case "failed":
        return "bg-red-100 text-red-800";
      default:
        return "bg-gray-100 text-gray-800";
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
          <span
            class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${this.getStatusBadgeClass(
              status.status || "unknown",
            )}"
          >
            ${status.status || "unknown"}
          </span>
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
            <div
              class="bg-white shadow rounded-lg border border-gray-200"
              data-testid="sync-source-${source.id || "unknown"}"
            >
              <div class="px-6 py-4 border-b border-gray-200">
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
            </div>
          `,
        )}
      </div>
    `;
  }

  render() {
    return html`
      <div class="max-w-6xl mx-auto px-4 py-8">
        <div class="mb-8">
          <h1
            class="text-3xl font-bold text-gray-900 mb-2"
            data-testid="page-title"
          >
            Settings
          </h1>
          <p class="text-gray-600" data-testid="page-description">
            Sync source configuration and status information
          </p>
        </div>

        ${this.isLoading
          ? html`
              <div class="flex justify-center items-center py-8">
                <div class="text-gray-500">Loading settings...</div>
              </div>
            `
          : this.error
            ? html`
                <div class="bg-red-50 border border-red-200 rounded-lg p-4">
                  <h2 class="text-red-800 font-semibold mb-2">
                    Error loading settings
                  </h2>
                  <p class="text-red-600">${this.error}</p>
                </div>
              `
            : this.renderSources()}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "settings-page": SettingsPage;
  }
}
