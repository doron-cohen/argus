import { LitElement, html } from "lit";
import { customElement, property } from "lit/decorators.js";
import type { Component } from "../../api/services/components/client";
import type { CheckReport } from "../../api/services/components/client";
import "../../ui/components/ui-badge.js";
import "../../ui/components/ui-card.js";
import "../../ui/components/ui-card-header.js";
import "../../ui/components/ui-info-row.js";
import "../../ui/components/ui-spinner.js";
import "../../ui/components/ui-alert.js";

import { nothing } from "lit";

@customElement("component-details")
export class ComponentDetails extends LitElement {
  @property({ type: Object, attribute: false })
  component: Component | null = null;

  @property({ type: Boolean, attribute: false })
  isLoading = false;

  @property({ type: String, attribute: false })
  errorMessage: string | null = null;

  @property({ type: Array, attribute: false })
  reports: readonly CheckReport[] = [];

  @property({ type: Boolean, attribute: false })
  isReportsLoading = false;

  @property({ type: String, attribute: false })
  reportsErrorMessage: string | null = null;

  private formatTimestamp(timestamp: string): string {
    return new Date(timestamp).toLocaleString();
  }

  render() {
    // Follow functional rendering model - return same result for same inputs
    if (this.isLoading) {
      return this.renderLoadingState();
    }

    if (this.errorMessage) {
      return this.renderErrorState();
    }

    if (!this.component) {
      return this.renderEmptyState();
    }

    return this.renderComponentDetails();
  }

  private renderLoadingState() {
    return html`
      <div
        class="u-flex u-justify-center u-items-center u-py-8"
        data-testid="component-details-loading"
      >
        <ui-spinner size="lg" color="primary"></ui-spinner>
        <span class="u-ml-2 u-text-muted">Loading component details...</span>
      </div>
    `;
  }

  private renderErrorState() {
    return html`
      <ui-card data-testid="component-details-error">
        <ui-alert
          variant="error"
          title="Error loading component"
          message="${this.errorMessage || ""}"
        >
        </ui-alert>
      </ui-card>
    `;
  }

  private renderEmptyState() {
    return html`
      <div class="u-text-center u-py-8 u-text-gray-500">
        No component data available
      </div>
    `;
  }

  private renderComponentDetails() {
    return html`
      <div data-testid="component-details">
        ${this.renderComponentHeader()} ${this.renderComponentInfo()}
        ${this.renderReports()}
      </div>
    `;
  }

  private renderComponentHeader() {
    return html`
      <ui-card>
        <ui-card-header
          title="${this.component!.name}"
          subtitle="ID: ${this.component!.id || this.component!.name}"
          title-data-testid="component-name"
          subtitle-data-testid="component-id"
        ></ui-card-header>
      </ui-card>
    `;
  }

  private renderComponentInfo() {
    return html`
      <ui-card class="u-mt-4">
        ${this.renderInfoRow(
          "Description",
          "description-label",
          "component-description",
          this.component!.description || "No description available"
        )}
        ${this.renderInfoRow(
          "Team",
          "team-label",
          "component-team",
          this.component!.owners?.team || "No team assigned"
        )}
        ${this.renderMaintainersRow()}
      </ui-card>
    `;
  }

  private renderInfoRow(
    label: string,
    labelTestId: string,
    valueTestId: string,
    value: string
  ) {
    return html`
      <ui-info-row
        label="${label}"
        value="${value}"
        label-data-testid="${labelTestId}"
        value-data-testid="${valueTestId}"
      ></ui-info-row>
    `;
  }

  private renderMaintainersRow() {
    const maintainers = this.component!.owners?.maintainers;
    if (!maintainers?.length) {
      return this.renderInfoRow(
        "Maintainers",
        "maintainers-label",
        "component-maintainers",
        "No maintainers assigned"
      );
    }

    return html`
      <ui-info-row
        label="Maintainers"
        value="${maintainers.join(", ")}"
        label-data-testid="maintainers-label"
        value-data-testid="component-maintainers"
      ></ui-info-row>
    `;
  }

  private renderReports() {
    if (this.isReportsLoading) {
      return this.renderReportsLoadingState();
    }

    if (this.reportsErrorMessage) {
      return this.renderReportsErrorState();
    }

    if (!this.reports || this.reports.length === 0) {
      return this.renderReportsEmptyState();
    }

    return this.renderReportsList();
  }

  private renderReportsLoadingState() {
    return html`
      <div class="u-mt-8">
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

  private renderReportsErrorState() {
    return html`
      <div class="u-mt-8" data-testid="reports-error">
        <ui-alert
          variant="error"
          title="Error loading reports"
          message="${this.reportsErrorMessage || ""}"
        >
        </ui-alert>
      </div>
    `;
  }

  private renderReportsEmptyState() {
    return html`
      <div
        class="u-mt-8 u-text-center u-text-gray-500"
        data-testid="no-reports"
      >
        No reports available for this component
      </div>
    `;
  }

  private renderReportsList() {
    return html`
      <div class="u-mt-8">
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
    "component-details": ComponentDetails;
  }
}
