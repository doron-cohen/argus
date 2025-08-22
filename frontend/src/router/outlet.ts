import { LitElement, html, type TemplateResult } from "lit";
import { customElement } from "lit/decorators.js";
import Navigo from "navigo";
import "../components/component-details";
import "../pages/home/index";
import "../pages/component-details/index";

@customElement("router-outlet")
export class RouterOutlet extends LitElement {
  private router: Navigo | null = null;
  private currentView: TemplateResult = html``;

  connectedCallback(): void {
    super.connectedCallback();
    this.initializeRouter();
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    // Clean up router when component is destroyed
    if (this.router) {
      this.router.destroy();
      this.router = null;
    }
  }

  private initializeRouter(): void {
    this.router = new Navigo("/");

    // Define routes
    this.router
      .on("/", () => {
        this.setView(html`<home-page></home-page>`);
      })
      .on("/components", () => {
        this.setView(html`<home-page></home-page>`);
      })
      .on("/components/:componentId", (match) => {
        // In Navigo v8, parameters are accessed via match.params
        const componentId = match?.params?.componentId || "";
        this.setView(
          html`<component-details-page
            component-id="${componentId}"
          ></component-details-page>`
        );
      })
      .notFound(() => {
        this.setView(html`<div>Page not found</div>`);
      });

    // Start the router
    this.router.resolve();
  }

  private setView(view: TemplateResult): void {
    this.currentView = view;
    this.requestUpdate();
  }

  render(): TemplateResult {
    return this.currentView;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "router-outlet": RouterOutlet;
  }
}
