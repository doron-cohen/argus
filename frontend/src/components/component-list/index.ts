import { LitElement, html } from "lit";
import { escapeHtml } from "../../utils";
import { getComponents } from "../../api/services/components/client";

export interface ComponentItem {
  id: string;
  name: string;
  description: string;
  owners: {
    maintainers: string[];
    team: string;
  };
}

export class ComponentList extends LitElement {
  private components: ComponentItem[] = [];
  private isLoading = true;
  private error: string | null = null;

  protected createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.loadComponents();
  }

  private async loadComponents(): Promise<void> {
    try {
      this.isLoading = true;
      this.error = null;
      this.requestUpdate();

      const { status, data } = await getComponents();
      const statusCode = typeof status === "number" ? status : 200;
      // Defensive: ensure we always have an array to render
      if (!Array.isArray(data as unknown as any[])) {
        this.components = [];
        this.error =
          statusCode >= 400 ? `HTTP ${statusCode}` : "Invalid API response";
      } else {
        this.components = data as unknown as ComponentItem[];
      }
    } catch (err) {
      this.error =
        err instanceof Error ? err.message : "Failed to fetch components";
      console.error("Error fetching components:", err);
    } finally {
      this.isLoading = false;
      this.requestUpdate();
      this.updateHeader();
    }
  }

  private updateHeader(): void {
    const header = document.querySelector('[data-testid="components-header"]');
    if (!header) return;
    if (this.error || this.isLoading) {
      header.textContent = "Components";
    } else {
      header.textContent = `Components (${this.components.length})`;
    }
  }

  render() {
    if (this.isLoading) {
      return html`
        <div class="overflow-x-auto">
          <table
            class="min-w-full divide-y divide-gray-200"
            data-testid="components-table"
          >
            <thead class="bg-gray-50">
              <tr>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-name"
                >
                  Name
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-id"
                >
                  ID
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-description"
                >
                  Description
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-team"
                >
                  Team
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-maintainers"
                >
                  Maintainers
                </th>
              </tr>
            </thead>
            <tbody
              class="bg-white divide-y divide-gray-200"
              data-testid="components-tbody"
            >
              <tr>
                <td colspan="5" class="px-6 py-4 text-center">
                  <div
                    class="text-sm text-gray-500"
                    data-testid="loading-message"
                  >
                    Loading components...
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      `;
    }

    if (this.error) {
      return html`
        <div class="overflow-x-auto">
          <table
            class="min-w-full divide-y divide-gray-200"
            data-testid="components-table"
          >
            <thead class="bg-gray-50">
              <tr>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-name"
                >
                  Name
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-id"
                >
                  ID
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-description"
                >
                  Description
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-team"
                >
                  Team
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-maintainers"
                >
                  Maintainers
                </th>
              </tr>
            </thead>
            <tbody
              class="bg-white divide-y divide-gray-200"
              data-testid="components-tbody"
            >
              <tr>
                <td colspan="5" class="px-6 py-4 text-center">
                  <div class="text-sm text-red-500" data-testid="error-message">
                    Error: ${escapeHtml(this.error)}
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      `;
    }

    if (this.components.length === 0) {
      return html`
        <div class="overflow-x-auto">
          <table
            class="min-w-full divide-y divide-gray-200"
            data-testid="components-table"
          >
            <thead class="bg-gray-50">
              <tr>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-name"
                >
                  Name
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-id"
                >
                  ID
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-description"
                >
                  Description
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-team"
                >
                  Team
                </th>
                <th
                  scope="col"
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  data-testid="header-maintainers"
                >
                  Maintainers
                </th>
              </tr>
            </thead>
            <tbody
              class="bg-white divide-y divide-gray-200"
              data-testid="components-tbody"
            >
              <tr>
                <td colspan="5" class="px-6 py-4 text-center">
                  <div
                    class="text-sm text-gray-500"
                    data-testid="no-components-message"
                  >
                    No components found
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      `;
    }

    return html`
      <div class="overflow-x-auto">
        <table
          class="min-w-full divide-y divide-gray-200"
          data-testid="components-table"
        >
          <thead class="bg-gray-50">
            <tr>
              <th
                scope="col"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                data-testid="header-name"
              >
                Name
              </th>
              <th
                scope="col"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                data-testid="header-id"
              >
                ID
              </th>
              <th
                scope="col"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                data-testid="header-description"
              >
                Description
              </th>
              <th
                scope="col"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                data-testid="header-team"
              >
                Team
              </th>
              <th
                scope="col"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                data-testid="header-maintainers"
              >
                Maintainers
              </th>
            </tr>
          </thead>
          <tbody
            class="bg-white divide-y divide-gray-200"
            data-testid="components-tbody"
          >
            ${this.components.map((comp) => {
              const slug = comp.id || comp.name;
              const href = `/components/${encodeURIComponent(slug)}`;
              return html`
                <tr
                  class="hover:bg-gray-50 cursor-pointer"
                  data-testid="component-row"
                  data-component-id="${slug}"
                >
                  <td class="px-6 py-4 whitespace-nowrap">
                    <a
                      href="${href}"
                      class="text-sm font-medium text-indigo-600 hover:text-indigo-900"
                      data-testid="component-name"
                    >
                      ${comp.name}
                    </a>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div
                      class="text-sm text-gray-500"
                      data-testid="component-id"
                    >
                      ${comp.id || comp.name}
                    </div>
                  </td>
                  <td class="px-6 py-4">
                    <div
                      class="text-sm text-gray-900"
                      data-testid="component-description"
                    >
                      ${comp.description || ""}
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div
                      class="text-sm text-gray-500"
                      data-testid="component-team"
                    >
                      ${comp.owners?.team || ""}
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div
                      class="text-sm text-gray-500"
                      data-testid="component-maintainers"
                    >
                      ${comp.owners?.maintainers?.join(", ") || ""}
                    </div>
                  </td>
                </tr>
              `;
            })}
          </tbody>
        </table>
      </div>
    `;
  }
}

if (!customElements.get("component-list")) {
  customElements.define("component-list", ComponentList);
}
