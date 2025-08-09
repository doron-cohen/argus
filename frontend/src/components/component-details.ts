import { BaseComponent } from "../utils/base-component";
import {
  componentDetails,
  loading,
  error,
  latestReports,
  reportsLoading,
  reportsError,
  type CheckReport,
} from "../stores/app-store";
import { escapeHtml } from "../utils";
import "../ui/ui-badge";

export class ComponentDetails extends BaseComponent {
  connectedCallback() {
    this.bindState(componentDetails, (component) => this.render(component));
    this.bindState(loading, (loading) => this.showLoading(loading));
    this.bindState(error, (error) => this.showError(error));
    this.bindState(latestReports, (reports) => this.renderReports(reports));
    this.bindState(reportsLoading, (loading) =>
      this.showReportsLoading(loading),
    );
    this.bindState(reportsError, (error) => this.showReportsError(error));
  }

  private render(component: any) {
    if (!component) {
      this.innerHTML = "";
      return;
    }

    // Pre-compute safe display texts for cleaner templates
    const vm = this.buildViewModel(component);

    this.innerHTML = `
      <div class="bg-white shadow overflow-hidden sm:rounded-lg" data-testid="component-details">
        <div class="px-4 py-5 sm:px-6">
          <h3 class="text-lg leading-6 font-medium text-gray-900" data-testid="component-name">${vm.nameText}</h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500" data-testid="component-id">ID: ${vm.idText}</p>
        </div>
        <div class="border-t border-gray-200">
          <dl>
            <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt class="text-sm font-medium text-gray-500" data-testid="description-label">Description</dt>
              <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2" data-testid="component-description">${vm.descriptionText}</dd>
            </div>
            <div class="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt class="text-sm font-medium text-gray-500" data-testid="team-label">Team</dt>
              <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2" data-testid="component-team">${vm.teamText}</dd>
            </div>
            <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt class="text-sm font-medium text-gray-500" data-testid="maintainers-label">Maintainers</dt>
              <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2" data-testid="component-maintainers">${vm.maintainersText}</dd>
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

    // Initialize reports section with current state
    this.renderReports(latestReports.get());
  }

  private buildViewModel(component: any) {
    const nameText = escapeHtml(component?.name ?? "");
    const idText = escapeHtml(component?.id ?? component?.name ?? "");
    const descriptionText = escapeHtml(
      component?.description || "No description available",
    );
    const teamText = escapeHtml(component?.owners?.team || "No team assigned");
    const maintainersText = escapeHtml(
      (component?.owners?.maintainers || []).join(", ") ||
        "No maintainers assigned",
    );
    return { nameText, idText, descriptionText, teamText, maintainersText };
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
        return `
        <div class="flex items-center justify-between py-2 border-b border-gray-100 last:border-b-0" data-testid="report-item">
          <span class="text-sm font-medium text-gray-900" data-testid="check-name">
            ${escapeHtml(report.check_slug)}
          </span>
          <div class="flex items-center space-x-2">
            <ui-badge status="${escapeHtml(
              report.status
            )}" data-testid="check-status"></ui-badge>
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

  // Status presentation is handled by <ui-badge>

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
        <div data-testid="component-details">
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
        </div>
      `;
    }
  }

  private showError(errorMessage: string | null) {
    if (errorMessage) {
      this.innerHTML = `
        <div data-testid="component-details">
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
                    <div class="mt-2 text-sm text-red-700" data-testid="error-message">${escapeHtml(
                      errorMessage,
                    )}</div>
                  </div>
                </div>
              </div>
            </div>
            <div class="px-4 py-3 bg-gray-50 text-right sm:px-6">
              <a href="/" class="text-sm font-medium text-indigo-600 hover:text-indigo-500" data-testid="back-to-components">← Back to Components</a>
            </div>
          </div>
        </div>
      `;
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
