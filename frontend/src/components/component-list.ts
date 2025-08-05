import { BaseComponent } from "../utils/base-component";
import { escapeHtml } from "../utils";

export interface Component {
  id: string;
  name: string;
  description: string;
  owners: {
    maintainers: string[];
    team: string;
  };
}

export class ComponentList extends BaseComponent {
  private components: Component[] = [];
  private isLoading = true;
  private error: string | null = null;

  connectedCallback() {
    this.loadComponents();
  }

  private async loadComponents(): Promise<void> {
    try {
      this.isLoading = true;
      this.error = null;
      this.render();

      const response = await fetch("/api/catalog/v1/components");

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.error || `HTTP ${response.status}: ${response.statusText}`
        );
      }

      this.components = await response.json();
    } catch (err) {
      this.error =
        err instanceof Error ? err.message : "Failed to fetch components";
      console.error("Error fetching components:", err);
    } finally {
      this.isLoading = false;
      this.render();
      this.updateHeader();
    }
  }

  private updateHeader(): void {
    const header = document.querySelector('[data-testid="components-header"]');
    if (header) {
      if (this.error) {
        header.textContent = "Components";
      } else if (this.isLoading) {
        header.textContent = "Components";
      } else {
        header.textContent = `Components (${this.components.length})`;
      }
    }
  }

  private render(): void {
    if (this.isLoading) {
      this.innerHTML = `
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200" data-testid="components-table">
            <thead class="bg-gray-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-name">Name</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-id">ID</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-description">Description</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-team">Team</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-maintainers">Maintainers</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200" data-testid="components-tbody">
              <tr>
                <td colspan="5" class="px-6 py-4 text-center">
                  <div class="text-sm text-gray-500" data-testid="loading-message">Loading components...</div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      `;
      return;
    }

    if (this.error) {
      this.innerHTML = `
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200" data-testid="components-table">
            <thead class="bg-gray-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-name">Name</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-id">ID</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-description">Description</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-team">Team</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-maintainers">Maintainers</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200" data-testid="components-tbody">
              <tr>
                <td colspan="5" class="px-6 py-4 text-center">
                  <div class="text-sm text-red-500" data-testid="error-message">Error: ${escapeHtml(
                    this.error
                  )}</div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      `;
      return;
    }

    if (this.components.length === 0) {
      this.innerHTML = `
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200" data-testid="components-table">
            <thead class="bg-gray-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-name">Name</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-id">ID</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-description">Description</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-team">Team</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-maintainers">Maintainers</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200" data-testid="components-tbody">
              <tr>
                <td colspan="5" class="px-6 py-4 text-center">
                  <div class="text-sm text-gray-500" data-testid="no-components-message">No components found</div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      `;
      return;
    }

    // Render table rows with clickable links
    this.innerHTML = `
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200" data-testid="components-table">
          <thead class="bg-gray-50">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-name">Name</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-id">ID</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-description">Description</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-team">Team</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider" data-testid="header-maintainers">Maintainers</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200" data-testid="components-tbody">
            ${this.components
              .map(
                (comp) => `
              <tr class="hover:bg-gray-50 cursor-pointer" data-testid="component-row" data-component-id="${escapeHtml(
                comp.id || comp.name
              )}">
                <td class="px-6 py-4 whitespace-nowrap">
                  <a href="/components/${escapeHtml(
                    comp.id || comp.name
                  )}" class="text-sm font-medium text-indigo-600 hover:text-indigo-900" data-testid="component-name">
                    ${escapeHtml(comp.name)}
                  </a>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-500" data-testid="component-id">${escapeHtml(
                    comp.id || comp.name
                  )}</div>
                </td>
                <td class="px-6 py-4">
                  <div class="text-sm text-gray-900" data-testid="component-description">${escapeHtml(
                    comp.description || ""
                  )}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-500" data-testid="component-team">${escapeHtml(
                    comp.owners?.team || ""
                  )}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-500" data-testid="component-maintainers">${escapeHtml(
                    comp.owners?.maintainers?.join(", ") || ""
                  )}</div>
                </td>
              </tr>
            `
              )
              .join("")}
          </tbody>
        </table>
      </div>
    `;
  }
}

// Register the web component
if (!customElements.get("component-list")) {
  customElements.define("component-list", ComponentList);
}
