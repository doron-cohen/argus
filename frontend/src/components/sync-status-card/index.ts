import { LitElement, html, nothing } from "lit";
import { customElement, property } from "lit/decorators.js";
import "../../ui/components/ui-badge.js";
import "../../ui/components/ui-loading-indicator.js";

export interface SyncStatusData {
  status?: string;
  lastSync?: string;
  componentsCount?: number;
  duration?: string;
  lastError?: string;
}

@customElement("sync-status-card")
export class SyncStatusCard extends LitElement {
  @property({ type: Object, attribute: false })
  status: SyncStatusData | null = null;

  @property({ type: Boolean, attribute: false })
  isLoading = false;

  @property({ type: String, attribute: false })
  error: string | null = null;

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

  render() {
    if (this.isLoading) {
      return html`
        <ui-loading-indicator
          message="Loading..."
          size="sm"
        ></ui-loading-indicator>
      `;
    }

    if (this.error) {
      return html`
        <div class="u-text-danger u-text-sm">
          <span class="u-font-medium">Error:</span> ${this.error}
        </div>
      `;
    }

    if (!this.status) {
      return html`
        <div class="u-text-muted u-text-sm">No status available</div>
      `;
    }

    // Handle case where status object exists but has no meaningful data
    if (
      this.status &&
      typeof this.status === "object" &&
      !this.status.status &&
      !this.status.lastSync &&
      !this.status.componentsCount
    ) {
      return html`
        <div class="u-text-muted u-text-sm">Status data loading...</div>
      `;
    }

    return html`
      <div class="u-stack-3">
        <div class="u-flex u-items-center u-gap-2">
          <ui-badge
            status=${this.getStatusBadgeStatus(this.status.status || "unknown")}
          >
            ${this.status.status || "unknown"}
          </ui-badge>
        </div>

        <div class="u-grid-2 u-text-sm">
          <div>
            <span class="u-font-medium u-text-secondary">Last Sync:</span>
            <div class="u-text-primary">
              ${this.formatTimestamp(this.status.lastSync)}
            </div>
          </div>
          <div>
            <span class="u-font-medium u-text-secondary">Components:</span>
            <div class="u-text-primary">${this.status.componentsCount}</div>
          </div>
          <div>
            <span class="u-font-medium u-text-secondary">Duration:</span>
            <div class="u-text-primary">
              ${this.formatDuration(this.status.duration)}
            </div>
          </div>
        </div>

        ${this.status.lastError
          ? html`
              <div class="u-mt-2">
                <span class="u-font-medium u-text-secondary">Last Error:</span>
                <div class="u-text-danger u-text-sm u-mt-1">
                  ${this.status.lastError}
                </div>
              </div>
            `
          : nothing}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "sync-status-card": SyncStatusCard;
  }
}

// Only define if not already defined (handles hot reload scenarios)
if (!customElements.get("sync-status-card")) {
  customElements.define("sync-status-card", SyncStatusCard);
}
