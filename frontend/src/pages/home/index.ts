import { LitElement, html } from "lit";
import "../../components/component-list/index";

export class HomePage extends LitElement {
  // Keep light DOM for global styles and tests
  protected createRenderRoot(): this {
    return this;
  }

  render() {
    return html`
      <div class="container mx-auto px-4 py-8">
        <div class="mb-8">
          <h1 class="text-3xl font-bold text-gray-900 mb-2" data-testid="page-title">Component Catalog</h1>
          <p class="text-gray-600" data-testid="page-description">Browse and search components in the Argus catalog</p>
        </div>

        <div data-testid="components-container">
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg leading-6 font-medium text-gray-900" data-testid="components-header">
              Components
            </h3>
          </div>
          <component-list></component-list>
        </div>
      </div>
    `;
  }
}

if (!customElements.get("home-page")) {
  customElements.define("home-page", HomePage);
}


