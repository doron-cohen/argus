import { LitElement, html } from "lit";
import { customElement, property } from "lit/decorators.js";
import type { Component } from "../../api/services/components/client";
import type { CheckReport } from "../../api/services/components/client";
import "../../ui/components/ui-badge.js";
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
        class="flex justify-center items-center py-8"
        data-testid="component-details-loading"
      >
        <div
          class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"
        ></div>
        <span class="ml-2 text-gray-600">Loading component details...</span>
      </div>
    `;
  }

  private renderErrorState() {
    return html`
      <div
        class="bg-white shadow overflow-hidden sm:rounded-lg"
        data-testid="component-details-error"
      >
        <div class="px-4 py-5 sm:px-6">
          <div class="rounded-md bg-red-50 p-4">
            <div class="flex">
              <div class="flex-shrink-0">
                <svg
                  class="h-5 w-5 text-red-400"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fill-rule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clip-rule="evenodd"
                  />
                </svg>
              </div>
              <div class="ml-3">
                <h3
                  class="text-sm font-medium text-red-800"
                  data-testid="error-title"
                >
                  Error loading component
                </h3>
                <div class="mt-2 text-sm text-red-700">
                  <p>${this.errorMessage}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    `;
  }

  private renderEmptyState() {
    return html`
      <div class="text-center py-8 text-gray-500">
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
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6">
          <h3
            class="text-lg leading-6 font-medium text-gray-900"
            data-testid="component-name"
          >
            ${this.component!.name}
          </h3>
          <p
            class="mt-1 max-w-2xl text-sm text-gray-500"
            data-testid="component-id"
          >
            ID: ${this.component!.id || this.component!.name}
          </p>
        </div>
      </div>
    `;
  }

  private renderComponentInfo() {
    return html`
      <div class="bg-white shadow overflow-hidden sm:rounded-lg mt-4">
        <div class="border-t border-gray-200">
          <dl>
            ${this.renderInfoRow(
              "Description",
              "description-label",
              "component-description",
              this.component!.description || "No description available",
            )}
            ${this.renderInfoRow(
              "Team",
              "team-label",
              "component-team",
              this.component!.owners?.team || "No team assigned",
            )}
            ${this.renderMaintainersRow()}
          </dl>
        </div>
      </div>
    `;
  }

  private renderInfoRow(
    label: string,
    labelTestId: string,
    valueTestId: string,
    value: string,
  ) {
    return html`
      <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
        <dt
          class="text-sm font-medium text-gray-500"
          data-testid="${labelTestId}"
        >
          ${label}
        </dt>
        <dd
          class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2"
          data-testid="${valueTestId}"
        >
          ${value}
        </dd>
      </div>
    `;
  }

  private renderMaintainersRow() {
    const maintainers = this.component!.owners?.maintainers;
    if (!maintainers?.length) {
      return this.renderInfoRow(
        "Maintainers",
        "maintainers-label",
        "component-maintainers",
        "No maintainers assigned",
      );
    }

    return html`
      <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
        <dt
          class="text-sm font-medium text-gray-500"
          data-testid="maintainers-label"
        >
          Maintainers
        </dt>
        <dd
          class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2"
          data-testid="component-maintainers"
        >
          ${maintainers.join(", ")}
        </dd>
      </div>
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
      <div class="mt-8">
        <div
          class="flex justify-center items-center py-4"
          data-testid="reports-loading"
        >
          <div
            class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"
          ></div>
          <span class="ml-2 text-gray-600">Loading reports...</span>
        </div>
      </div>
    `;
  }

  private renderReportsErrorState() {
    return html`
      <div class="mt-8">
        <div class="rounded-md bg-red-50 p-4" data-testid="reports-error">
          <div class="flex">
            <div class="flex-shrink-0">
              <svg
                class="h-5 w-5 text-red-400"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fill-rule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                  clip-rule="evenodd"
                />
              </svg>
            </div>
            <div class="ml-3">
              <h3 class="text-sm font-medium text-red-800">
                Error loading reports
              </h3>
              <div class="mt-2 text-sm text-red-700">
                <p>${this.reportsErrorMessage}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    `;
  }

  private renderReportsEmptyState() {
    return html`
      <div class="mt-8 text-center text-gray-500" data-testid="no-reports">
        No reports available for this component
      </div>
    `;
  }

  private renderReportsList() {
    return html`
      <div class="mt-8">
        <h3
          class="text-lg font-semibold text-gray-900 mb-4"
          data-testid="reports-label"
        >
          Latest Quality Checks
        </h3>
        <div class="space-y-4" data-testid="reports-list">
          ${this.reports.map((report) => this.renderReportItem(report))}
        </div>
      </div>
    `;
  }

  private renderReportItem(report: CheckReport) {
    return html`
      <div
        class="bg-white border border-gray-200 rounded-lg p-4"
        data-testid="report-item"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <span
              class="text-sm font-medium text-gray-900"
              data-testid="check-name"
            >
              ${report.check_slug}
            </span>
            <ui-badge
              status="${report.status}"
              data-testid="check-status"
            ></ui-badge>
          </div>
          <span class="text-sm text-gray-500" data-testid="check-timestamp">
            ${this.formatTimestamp(report.timestamp)}
          </span>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "component-details": ComponentDetails;
  }
}
