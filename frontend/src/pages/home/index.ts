import { LitElement, html } from "lit";
import { customElement, state } from "lit/decorators.js";
import {
  getComponents,
  type Component,
} from "../../api/services/components/client";
import "../../components/component-list/index";
import "../../ui/primitives/page-container.js";
import "../../ui/components/ui-page-header.js";
import "../../ui/primitives/ui-stack.js";

@customElement("home-page")
export class HomePage extends LitElement {
  @state()
  components: Component[] = [];

  @state()
  isLoading = true;

  @state()
  error: string | null = null;

  async connectedCallback(): Promise<void> {
    super.connectedCallback();
    await this.loadComponents();
  }

  private async loadComponents(): Promise<void> {
    try {
      this.isLoading = true;
      this.error = null;

      const response = await getComponents();
      const statusCode =
        typeof response.status === "number" ? response.status : 200;
      const componentsData = response.data;

      // Defensive: ensure we always have an array to render
      if (!Array.isArray(componentsData)) {
        this.components = [];
        this.error =
          statusCode >= 400 ? `HTTP ${statusCode}` : "Invalid API response";
      } else if (statusCode < 200 || statusCode >= 300) {
        throw new Error(`HTTP ${statusCode}`);
      } else {
        this.components = componentsData;
      }
    } catch (err) {
      this.error =
        err instanceof Error ? err.message : "Failed to fetch components";
      console.error("Error fetching components:", err);
    } finally {
      this.isLoading = false;
    }
  }

  render() {
    return html`
      <ui-page-container max-width="xl" padding="lg">
        <ui-page-header
          title="Component Catalog"
          description="Browse and search components in the Argus catalog"
          size="lg"
        ></ui-page-header>

        <main>
          <ui-stack gap="lg">
            <div data-testid="components-container">
              <h2 class="u-section-title" data-testid="components-header">
                Components${this.isLoading || this.error
                  ? ""
                  : ` (${this.components.length})`}
              </h2>
              <component-list
                .components=${this.components}
                .isLoading=${this.isLoading}
                .error=${this.error}
                id="component-list"
              ></component-list>
            </div>
          </ui-stack>
        </main>
      </ui-page-container>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "home-page": HomePage;
  }
}

// Only define if not already defined (handles hot reload scenarios)
if (!customElements.get("home-page")) {
  customElements.define("home-page", HomePage);
}
