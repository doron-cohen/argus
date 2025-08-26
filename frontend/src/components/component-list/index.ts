import { LitElement, html } from "lit";
import { customElement, property } from "lit/decorators.js";
import { nothing } from "lit";
import { escapeHtml } from "../../utils";
import type { Component } from "../../api/services/components/client";
import "../../ui/components/ui-table.js";

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
      <ui-table data-testid="components-table">
        ${this.renderTableHeader()} ${this.renderTableBody()}
      </ui-table>
    `;
  }

  private renderTableHeader() {
    return html`
      <thead>
        <tr>
          <th scope="col" data-testid="header-name">Name</th>
          <th scope="col" data-testid="header-id">ID</th>
          <th scope="col" data-testid="header-description">Description</th>
          <th scope="col" data-testid="header-team">Team</th>
          <th scope="col" data-testid="header-maintainers">Maintainers</th>
        </tr>
      </thead>
    `;
  }

  private renderTableBody() {
    return html`
      <tbody data-testid="components-tbody">
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
        <td colspan="5" class="text-center">
          <div class="u-text-sm u-text-muted" data-testid="loading-message">
            Loading components...
          </div>
        </td>
      </tr>
    `;
  }

  private renderErrorRow() {
    return html`
      <tr>
        <td colspan="5" class="text-center">
          <div class="u-text-sm u-text-danger" data-testid="error-message">
            Error: ${escapeHtml(this.error!)}
          </div>
        </td>
      </tr>
    `;
  }

  private renderEmptyRow() {
    return html`
      <tr>
        <td colspan="5" class="text-center">
          <div
            class="u-text-sm u-text-muted"
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
      <tr data-testid="component-row" data-component-id="${slug}">
        <td class="whitespace-nowrap">
          <a
            href="${href}"
            class="u-text-sm u-font-medium u-text-primary hover:u-text-primary"
            data-testid="component-name"
          >
            ${comp.name}
          </a>
        </td>
        <td class="whitespace-nowrap">
          <div class="u-text-sm u-text-muted" data-testid="component-id">
            ${comp.id || comp.name}
          </div>
        </td>
        <td>
          <div
            class="u-text-sm u-text-primary"
            data-testid="component-description"
          >
            ${comp.description || ""}
          </div>
        </td>
        <td class="whitespace-nowrap">
          <div class="u-text-sm u-text-muted" data-testid="component-team">
            ${comp.owners?.team || ""}
          </div>
        </td>
        <td class="whitespace-nowrap">
          <div
            class="u-text-sm u-text-muted"
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
