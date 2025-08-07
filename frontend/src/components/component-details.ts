import { BaseComponent } from "../utils/base-component.ts";
import {
  componentDetails,
  loading,
  error,
  latestReports,
  reportsLoading,
  reportsError,
  type CheckReport,
} from "../stores/app-store.ts";
import { escapeHtml } from "../utils.ts";

export class ComponentDetails extends BaseComponent {
  connectedCallback() {
    this.bindState(componentDetails, (component) => this.render(component));
    this.bindState(loading, (loading) => this.showLoading(loading));
    this.bindState(error, (error) => this.showError(error));
    this.bindState(latestReports, (reports) => this.renderReports(reports));
    this.bindState(reportsLoading, (loading) =>
      this.showReportsLoading(loading)
    );
    this.bindState(reportsError, (error) => this.showReportsError(error));
  }

  private render(component: any) {
    if (!component) {
      this.innerHTML = "";
      return;
    }

    this.innerHTML = `
      <div class="bg-white shadow overflow-hidden sm:rounded-lg" data-testid="component-details">
        <div class="px-4 py-5 sm:px-6">
          <h3 class="text-lg leading-6 font-medium text-gray-900" data-testid="component-name"></h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500" data-testid="component-id"></p>
        </div>
        <div class="border-t border-gray-200">
          <dl>
            <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt class="text-sm font-medium text-gray-500" data-testid="description-label">Description</dt>
              <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2" data-testid="component-description"></dd>
            </div>
            <div class="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt class="text-sm font-medium text-gray-500" data-testid="team-label">Team</dt>
              <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2" data-testid="component-team"></dd>
            </div>
            <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt class="text-sm font-medium text-gray-500" data-testid="maintainers-label">Maintainers</dt>
              <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2" data-testid="component-maintainers"></dd>
            </div>
            <div class="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt class="text-sm font-medium text-gray-500" data-testid="reports-label">Latest Quality Checks</dt>
              <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2" data-testid="component-reports">
                <div id="reports-container"></div>
              </dd>
            </div>
          </dl>
        </div>
        <div class="px-4 py-3 bg-gray-50 text-right sm:px-6">
          <a href="/" class="text-sm font-medium text-indigo-600 hover:text-indigo-500" data-testid="back-to-components">
            ← Back to Components
          </a>
        </div>
      </div>
    `;

    // Assign safe text content and initialize reports
    this.setSafeTextContent(component);
    // Initialize reports section with current state
    this.renderReports(latestReports.get());
  }

  // Safely assign text content to avoid XSS
  private setSafeTextContent(component: any) {
    const nameElement = this.querySelector('[data-testid="component-name"]');
    const idElement = this.querySelector('[data-testid="component-id"]');
    const descriptionElement = this.querySelector('[data-testid="component-description"]');
    const teamElement = this.querySelector('[data-testid="component-team"]');
    const maintainersElement = this.querySelector('[data-testid="component-maintainers"]');

    if (nameElement) nameElement.textContent = component.name ?? "";
    if (idElement) idElement.textContent = `ID: ${component.id || component.name}`;
    if (descriptionElement)
      descriptionElement.textContent = component.description || "No description available";
    if (teamElement) teamElement.textContent = component.owners?.team || "No team assigned";
    if (maintainersElement)
      maintainersElement.textContent = (component.owners?.maintainers?.join(", ") || "No maintainers assigned");
  }

  private renderReports(reports: readonly CheckReport[]) {
    const container = this.querySelector("#reports-container");
    if (!container) return;

    if (reports.length === 0) {
      container.innerHTML = `
        <div class="text-gray-500 italic" data-testid="no-reports">
          No quality checks available
        </div>
      `;
      return;
    }

    const reportsList = reports
      .map((report) => {
        const statusClass = this.getStatusClass(report.status);
        const statusIcon = this.getStatusIcon(report.status);

        return `
        <div class="flex items-center justify-between py-2 border-b border-gray-100 last:border-b-0" data-testid="report-item">
          <span class="text-sm font-medium text-gray-900" data-testid="check-name">
            ${escapeHtml(report.check_slug)}
          </span>
          <div class="flex items-center space-x-2">
            <span class="${statusClass} px-2 py-1 rounded-full text-xs font-medium flex items-center" data-testid="check-status">
              ${statusIcon}
              ${escapeHtml(report.status)}
            </span>
            <span class="text-xs text-gray-500" data-testid="check-timestamp">
              ${this.formatTimestamp(report.timestamp)}
            </span>
          </div>
        </div>
      `;
      })
      .join("");

    container.innerHTML = `
      <div class="space-y-1" data-testid="reports-list">
        ${reportsList}
      </div>
    `;
  }

  private getStatusClass(status: string): string {
    switch (status) {
      case "pass":
        return "bg-green-100 text-green-800";
      case "fail":
      case "error":
      case "unknown":
        return "bg-red-100 text-red-800";
      case "disabled":
      case "skipped":
        return "bg-yellow-100 text-yellow-800";
      case "completed":
        return "bg-blue-100 text-blue-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  }

  private getStatusIcon(status: string): string {
    switch (status) {
      case "pass":
        return '<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"></path></svg>';
      case "fail":
      case "error":
      case "unknown":
        return '<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path></svg>';
      case "disabled":
      case "skipped":
        return '<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path></svg>';
      case "completed":
        return '<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path></svg>';
      default:
        return "";
    }
  }

  private formatTimestamp(timestamp: string): string {
    try {
      const date = new Date(timestamp);
      if (isNaN(date.getTime())) {
        return escapeHtml(String(timestamp));
      }
      return (
        date.toLocaleDateString() +
        " " +
        date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
      );
    } catch {
      return escapeHtml(String(timestamp));
    }
  }

  private showLoading(isLoading: boolean) {
    if (isLoading) {
      this.innerHTML = `
        <div class="bg-white shadow overflow-hidden sm:rounded-lg" data-testid="component-details-loading">
          <div class="px-4 py-5 sm:px-6">
            <div class="animate-pulse">
              <div class="h-4 bg-gray-200 rounded w-1/3 mb-2"></div>
              <div class="h-3 bg-gray-200 rounded w-1/4"></div>
            </div>
          </div>
          <div class="border-t border-gray-200">
            <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <div class="h-3 bg-gray-200 rounded w-1/4"></div>
              <div class="mt-1 sm:mt-0 sm:col-span-2">
                <div class="h-3 bg-gray-200 rounded w-full"></div>
              </div>
            </div>
            <div class="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <div class="h-3 bg-gray-200 rounded w-1/4"></div>
              <div class="mt-1 sm:mt-0 sm:col-span-2">
                <div class="h-3 bg-gray-200 rounded w-1/2"></div>
              </div>
            </div>
            <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <div class="h-3 bg-gray-200 rounded w-1/4"></div>
              <div class="mt-1 sm:mt-0 sm:col-span-2">
                <div class="h-3 bg-gray-200 rounded w-2/3"></div>
              </div>
            </div>
          </div>
        </div>
      `;
    }
  }

  private showError(errorMessage: string | null) {
    if (errorMessage) {
      this.innerHTML = `
        <div class="bg-white shadow overflow-hidden sm:rounded-lg" data-testid="component-details-error">
          <div class="px-4 py-5 sm:px-6">
            <div class="rounded-md bg-red-50 p-4">
              <div class="flex">
                <div class="flex-shrink-0">
                  <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                  </svg>
                </div>
                <div class="ml-3">
                  <h3 class="text-sm font-medium text-red-800" data-testid="error-title">Error loading component</h3>
                  <div class="mt-2 text-sm text-red-700" data-testid="error-message"></div>
                </div>
              </div>
            </div>
          </div>
          <div class="px-4 py-3 bg-gray-50 text-right sm:px-6">
            <a href="/" class="text-sm font-medium text-indigo-600 hover:text-indigo-500" data-testid="back-to-components">
              ← Back to Components
            </a>
          </div>
        </div>
      `;

      const errorMessageElement = this.querySelector('[data-testid="error-message"]');
      if (errorMessageElement) errorMessageElement.textContent = errorMessage;
    }
  }

  private showReportsLoading(loading: boolean) {
    const container = this.querySelector("#reports-container");
    if (!container) return;

    if (loading) {
      container.innerHTML = `
        <div class="text-gray-500 text-sm" data-testid="reports-loading">
          Loading quality checks...
        </div>
      `;
    }
  }

  private showReportsError(error: string | null) {
    const container = this.querySelector("#reports-container");
    if (!container) return;

    if (error) {
      container.innerHTML = `
        <div class="text-red-600 text-sm" data-testid="reports-error">
          Error loading quality checks: ${escapeHtml(error)}
        </div>
      `;
    }
  }
}

customElements.define("component-details", ComponentDetails);
