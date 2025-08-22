import { LitElement, html } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { loadComponentDetails, loadComponentReports } from "./data";
import { resetComponentDetails, resetReports } from "./store";
import "../../components/component-details";
import {
  componentDetails,
  loading,
  error as errorStore,
  latestReports,
  reportsLoading,
  reportsError,
  type Component,
  type CheckReport,
} from "./store";

@customElement("component-details-page")
export class ComponentDetailsPage extends LitElement {
  @property({ type: String, attribute: "component-id" })
  componentId = "";

  @state()
  component: Component | null = null;

  @state()
  isLoading = false;

  @state()
  errorMessage: string | null = null;

  @state()
  reports: readonly CheckReport[] = [];

  @state()
  isReportsLoading = false;

  @state()
  reportsErrorMessage: string | null = null;

  private unsubscribers: Array<() => void> = [];

  async connectedCallback(): Promise<void> {
    super.connectedCallback();

    // Initialize componentId from attribute or URL
    if (!this.componentId) {
      const fromAttr = this.getAttribute("component-id") || "";
      if (fromAttr) {
        this.componentId = fromAttr;
      } else if (typeof location !== "undefined") {
        const parts = location.pathname.split("/").filter(Boolean);
        if (parts[0] === "components" && parts[1]) {
          this.componentId = parts[1];
        }
      }
    }

    // Subscribe to stores
    this.unsubscribers.push(
      componentDetails.subscribe((value) => {
        this.component = value;
        this.requestUpdate();
      })
    );

    this.unsubscribers.push(
      loading.subscribe((value) => {
        this.isLoading = value;
        this.requestUpdate();
      })
    );

    this.unsubscribers.push(
      errorStore.subscribe((value) => {
        this.errorMessage = value;
        this.requestUpdate();
      })
    );

    this.unsubscribers.push(
      latestReports.subscribe((value) => {
        this.reports = value;
        this.requestUpdate();
      })
    );

    this.unsubscribers.push(
      reportsLoading.subscribe((value) => {
        this.isReportsLoading = value;
        this.requestUpdate();
      })
    );

    this.unsubscribers.push(
      reportsError.subscribe((value) => {
        this.reportsErrorMessage = value;
        this.requestUpdate();
      })
    );

    if (this.componentId) {
      void this.load();
    }
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.unsubscribers.forEach((unsubscribe) => unsubscribe());
    this.unsubscribers = [];
  }

  protected willUpdate(changed: Map<string, unknown>): void {
    // Trigger load as soon as componentId becomes available
    if (changed.has("componentId") && this.componentId) {
      void this.load();
    }
  }

  private async load(): Promise<void> {
    try {
      resetComponentDetails();
      resetReports();
      if (!this.componentId) return;

      await loadComponentDetails(this.componentId);

      if (!errorStore.get()) {
        await loadComponentReports(this.componentId);
      }
    } catch (err) {
      console.error("[ComponentDetailsPage] Error loading data:", err);
    }
  }

  render() {
    return html`
      <div class="container mx-auto px-4 py-8">
        ${this.renderPageHeader()} ${this.renderComponentDetails()}
      </div>
    `;
  }

  private renderPageHeader() {
    return html`
      <div class="mb-8">
        <div class="flex items-center justify-between">
          <div>
            <h1
              class="text-3xl font-bold text-gray-900 mb-2"
              data-testid="page-title"
            >
              Component Details
            </h1>
            <p class="text-gray-600" data-testid="page-description">
              View detailed information about the component
            </p>
          </div>
          ${this.renderBackButton()}
        </div>
      </div>
    `;
  }

  private renderBackButton() {
    return html`
      <a
        href="/components"
        class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
        data-testid="back-to-components"
      >
        ← Back to Components
      </a>
    `;
  }

  private renderComponentDetails() {
    return html`
      <component-details
        .component=${this.component}
        .isLoading=${this.isLoading}
        .errorMessage=${this.errorMessage}
        .reports=${this.reports}
        .isReportsLoading=${this.isReportsLoading}
        .reportsErrorMessage=${this.reportsErrorMessage}
      ></component-details>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "component-details-page": ComponentDetailsPage;
  }
}
