import { LitElement, html, type TemplateResult } from "lit";
import Navigo from "navigo";
import "../components/component-details";
import "../pages/home/index";
import { resetComponentDetails, resetReports, error } from "../stores/app-store";
import { loadComponentDetails, loadComponentReports } from "../pages/component-details/data";

export class RouterOutlet extends LitElement {
  private router: Navigo | null = null;
  private currentView: TemplateResult = html``;

  // Render into light DOM to keep global styles and testing selectors
  protected createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    if (!this.router) {
      this.router = new Navigo("/");
      this.router
        .on("/", () => this.setView(html`<home-page></home-page>`))
        .on("/components", () => this.setView(html`<home-page></home-page>`))
        .on("/components/:id", async (match) => {
          this.setView(html`
            <div class="container mx-auto px-4 py-8">
              <div class="mb-8">
                <h1 class="text-3xl font-bold text-gray-900 mb-2" data-testid="page-title">Component Details</h1>
                <p class="text-gray-600" data-testid="page-description">View detailed information about the component</p>
              </div>
              <component-details></component-details>
            </div>
          `);

          const componentId = match?.data?.id as string | undefined;
          if (!componentId) return;

          // Reset state and load data as in previous implementation
          resetComponentDetails();
          resetReports();

          await loadComponentDetails(componentId);
          if (!error.get()) {
            await loadComponentReports(componentId);
          }
        })
        .notFound(() => this.setView(html`
          <div class="container mx-auto px-4 py-8">
            <h1 class="text-2xl font-bold text-gray-900 mb-4">Page not found</h1>
            <a href="/" class="text-indigo-600 hover:text-indigo-500">Go home</a>
          </div>
        `))
        .resolve();
    }
  }

  private setView(view: TemplateResult): void {
    this.currentView = view;
    this.requestUpdate();
  }

  render() {
    return this.currentView;
  }
}

if (!customElements.get("router-outlet")) {
  customElements.define("router-outlet", RouterOutlet);
}


