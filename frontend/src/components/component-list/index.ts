import { LitElement, html } from "lit";
import { customElement, property } from "lit/decorators.js";
import { nothing } from "lit";
import { escapeHtml } from "../../utils";
import type { Component } from "../../api/services/components/client";

@customElement("component-list")
export class ComponentList extends LitElement {
  @property({ type: Array, attribute: false })
  components: Component[] = [];

  @property({ type: Boolean, attribute: false })
  isLoading = false;

  @property({ type: String, attribute: false })
  error: string | null = null;

  render() {
    return html`
      <div class="overflow-x-auto">
        <table
          class="min-w-full divide-y divide-gray-200"
          data-testid="components-table"
        >
          ${this.renderTableHeader()} ${this.renderTableBody()}
        </table>
      </div>
    `;
  }

  private renderTableHeader() {
    return html`
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
    `;
  }

  private renderTableBody() {
    return html`
      <tbody
        class="bg-white divide-y divide-gray-200"
        data-testid="components-tbody"
      >
        ${this.renderTableContent()}
      </tbody>
    `;
  }

  private renderTableContent() {
    if (this.isLoading) {
      return this.renderLoadingRow();
    }

    if (this.error) {
      return this.renderErrorRow();
    }

    if (this.components.length === 0) {
      return this.renderEmptyRow();
    }

    return this.components.map((comp) => this.renderComponentRow(comp));
  }

  private renderLoadingRow() {
    return html`
      <tr>
        <td colspan="5" class="px-6 py-4 text-center">
          <div class="text-sm text-gray-500" data-testid="loading-message">
            Loading components...
          </div>
        </td>
      </tr>
    `;
  }

  private renderErrorRow() {
    return html`
      <tr>
        <td colspan="5" class="px-6 py-4 text-center">
          <div class="text-sm text-red-500" data-testid="error-message">
            Error: ${escapeHtml(this.error!)}
          </div>
        </td>
      </tr>
    `;
  }

  private renderEmptyRow() {
    return html`
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
    `;
  }

  private renderComponentRow(comp: Component) {
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
          <div class="text-sm text-gray-500" data-testid="component-id">
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
          <div class="text-sm text-gray-500" data-testid="component-team">
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
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "component-list": ComponentList;
  }
}
