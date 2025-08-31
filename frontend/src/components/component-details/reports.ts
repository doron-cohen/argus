import { LitElement, html } from "lit";
import { customElement, property } from "lit/decorators.js";
import type { CheckReport } from "../../api/services/components/client";
import "../../ui/components/ui-card.js";
import "../../ui/components/ui-badge.js";
import "../../ui/components/ui-spinner.js";
import "../../ui/components/ui-alert.js";

@customElement("component-reports")
export class ComponentReports extends LitElement {
  @property({ type: Array, attribute: false })
  reports: readonly CheckReport[] = [];

  @property({ type: Boolean, attribute: false })
  isLoading = false;

  @property({ type: String, attribute: false })
  errorMessage: string | null = null;

  private formatTimestamp(timestamp: string): string {
    return new Date(timestamp).toLocaleString();
  }

  render() {
    if (this.isLoading) {
      return this.renderLoadingState();
    }

    if (this.errorMessage) {
      return this.renderErrorState();
    }

    if (!this.reports || this.reports.length === 0) {
      return this.renderEmptyState();
    }

    return this.renderReportsList();
  }

  private renderLoadingState() {
    return html`
      <div>
        <div
          class="u-flex u-justify-center u-items-center u-py-4"
          data-testid="reports-loading"
        >
          <ui-spinner size="md" color="primary"></ui-spinner>
          <span class="u-ml-2 u-text-muted">Loading reports...</span>
        </div>
      </div>
    `;
  }

  private renderErrorState() {
    return html`
      <div data-testid="reports-error">
        <ui-alert
          variant="error"
          title="Error loading reports"
          message="${this.errorMessage || ""}"
        >
        </ui-alert>
      </div>
    `;
  }

  private renderEmptyState() {
    return html`
      <div class="u-text-center u-text-gray-500" data-testid="no-reports">
        No reports available for this component
      </div>
    `;
  }

  private renderReportsList() {
    return html`
      <div>
        <h3
          class="u-text-lg u-font-semibold u-text-gray-900 u-mb-4"
          data-testid="reports-label"
        >
          Latest Quality Checks
        </h3>
        <div class="u-space-y-4" data-testid="reports-list">
          ${this.reports.map((report) => this.renderReportItem(report))}
        </div>
      </div>
    `;
  }

  private renderReportItem(report: CheckReport) {
    return html`
      <ui-card data-testid="report-item">
        <div class="u-flex u-items-center u-justify-between">
          <div class="u-flex u-items-center u-gap-3">
            <span
              class="u-text-sm u-font-medium u-text-gray-900"
              data-testid="check-name"
            >
              ${report.check_slug}
            </span>
            <ui-badge
              status="${report.status}"
              data-testid="check-status"
            ></ui-badge>
          </div>
          <span class="u-text-sm u-text-gray-500" data-testid="check-timestamp">
            ${this.formatTimestamp(report.timestamp)}
          </span>
        </div>
      </ui-card>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "component-reports": ComponentReports;
  }
}
