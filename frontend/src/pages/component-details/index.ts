import { LitElement, html } from "lit";
import { loadComponentDetails, loadComponentReports } from "./data";
import { resetComponentDetails, resetReports, error } from "./store";
import "../../components/component-details";
import {
  componentDetails,
  loading,
  error as errorStore,
  latestReports,
  reportsLoading,
  reportsError,
} from "./store";

export class ComponentDetailsPage extends LitElement {
  static properties = {
    componentId: { type: String, attribute: "component-id" },
  };

  componentId = "";

  protected createRenderRoot(): this {
    return this;
  }

  async connectedCallback(): Promise<void> {
    super.connectedCallback();
    // Reset and load data when mounted or when componentId changes
    await this.load();
  }

  async updated(prev: Map<string, unknown>): Promise<void> {
    if (
      prev.has("componentId") &&
      prev.get("componentId") !== this.componentId
    ) {
      await this.load();
    }
  }

  private async load(): Promise<void> {
    resetComponentDetails();
    resetReports();
    if (!this.componentId) return;
    await loadComponentDetails(this.componentId);
    if (!error.get()) {
      await loadComponentReports(this.componentId);
    }
  }

  render() {
    return html`
      <div class="container mx-auto px-4 py-8">
        <div class="mb-8">
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
        <component-details
          .component=${componentDetails.get()}
          .isLoading=${loading.get()}
          .errorMessage=${errorStore.get()}
          .reports=${latestReports.get()}
          .isReportsLoading=${reportsLoading.get()}
          .reportsErrorMessage=${reportsError.get()}
        ></component-details>
      </div>
    `;
  }
}

if (!customElements.get("component-details-page")) {
  customElements.define("component-details-page", ComponentDetailsPage);
}
