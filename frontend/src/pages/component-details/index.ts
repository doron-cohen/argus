import { LitElement, html } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { loadComponentDetails, loadComponentReports } from "./data";
import { resetComponentDetails, resetReports } from "./store";
import "../../components/component-details";
import "../../ui/primitives/page-container.js";
import "../../ui/components/ui-page-header.js";
import "../../ui/primitives/ui-stack.js";
import "../../ui/primitives/ui-cluster.js";
import "../../ui/components/ui-button.js";
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
      }),
    );

    this.unsubscribers.push(
      loading.subscribe((value) => {
        this.isLoading = value;
        this.requestUpdate();
      }),
    );

    this.unsubscribers.push(
      errorStore.subscribe((value) => {
        this.errorMessage = value;
        this.requestUpdate();
      }),
    );

    this.unsubscribers.push(
      latestReports.subscribe((value) => {
        this.reports = value;
        this.requestUpdate();
      }),
    );

    this.unsubscribers.push(
      reportsLoading.subscribe((value) => {
        this.isReportsLoading = value;
        this.requestUpdate();
      }),
    );

    this.unsubscribers.push(
      reportsError.subscribe((value) => {
        this.reportsErrorMessage = value;
        this.requestUpdate();
      }),
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
      <ui-page-container max-width="xl" padding="lg">
        ${this.renderPageHeader()} ${this.renderComponentDetails()}
      </ui-page-container>
    `;
  }

  private renderPageHeader() {
    return html`
      <ui-page-header
        title="Component Details"
        description="View detailed information about the component"
        size="lg"
      >
        <a
          slot="actions"
          href="/components"
          class="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          data-testid="back-to-components"
        >
          ← Back to Components
        </a>
      </ui-page-header>
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

// Only define if not already defined (handles hot reload scenarios)
if (!customElements.get("component-details-page")) {
  customElements.define("component-details-page", ComponentDetailsPage);
}
