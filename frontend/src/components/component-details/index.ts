import { LitElement, html } from "lit";
import { customElement, property } from "lit/decorators.js";
import type { Component } from "../../api/services/components/client";
import type { CheckReport } from "../../api/services/components/client";
import "../../ui/components/ui-spinner.js";
import "../../ui/components/ui-alert.js";
import "./metadata.js";
import "./reports.js";

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
        <component-metadata .component=${this.component}></component-metadata>
        <div class="u-mt-8">
          <component-reports
            .reports=${this.reports}
            .isLoading=${this.isReportsLoading}
            .errorMessage=${this.reportsErrorMessage}
          ></component-reports>
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
