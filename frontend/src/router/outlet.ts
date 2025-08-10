import { LitElement, html, type TemplateResult } from "lit";
import Navigo from "navigo";
import "../components/component-details";
import "../pages/home/index";
import "../pages/component-details/index";

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
        .on("/components/:id", (match) => {
          const componentId = (match?.data?.id as string) || "";
          this.setView(
            html`<component-details-page
              component-id="${componentId}"
            ></component-details-page>`
          );
        })
        .notFound(() =>
          this.setView(html`
            <div class="container mx-auto px-4 py-8">
              <h1 class="text-2xl font-bold text-gray-900 mb-4">
                Page not found
              </h1>
              <a href="/" class="text-indigo-600 hover:text-indigo-500"
                >Go home</a
              >
            </div>
          `)
        )
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
