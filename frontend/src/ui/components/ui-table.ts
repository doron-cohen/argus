import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-table")
export class UiTable extends LitElement {
  @property({ type: Boolean, reflect: true })
  striped = false;

  @property({ type: String, reflect: true })
  size: "sm" | "md" = "md";

  @property({ type: Boolean, reflect: true })
  loading = false;

  @property({ type: String, attribute: "loading-message" })
  loadingMessage = "Loading...";

  @property({ type: String, attribute: "error-message" })
  errorMessage: string | null = null;

  @property({ type: Boolean, reflect: true })
  empty = false;

  @property({ type: String, attribute: "empty-message" })
  emptyMessage = "No data found";

  @property({ type: Number, attribute: "col-span" })
  colSpan = 1;

  static styles = css`
    :host {
      display: block;
      width: 100%;
    }

    .table-container {
      overflow-x: auto;
    }

    .table {
      width: 100%;
      border-collapse: collapse;
      border-spacing: 0;
    }

    /* Header styling */
    .table thead {
      background-color: var(--color-bg-subtle, rgb(249 250 251));
    }

    .table thead th {
      padding: var(--space-3, 0.75rem) var(--space-6, 1.5rem);
      text-align: left;
      font-size: var(--font-size-xs, 0.75rem);
      font-weight: var(--font-weight-medium, 500);
      color: var(--color-fg-muted, rgb(107 114 128));
      text-transform: uppercase;
      letter-spacing: 0.05em;
      border-bottom: 1px solid var(--color-border, rgb(229 231 235));
    }

    /* Body styling */
    .table tbody {
      background-color: var(--color-bg, rgb(255 255 255));
    }

    .table tbody tr {
      border-bottom: 1px solid var(--color-border, rgb(229 231 235));
    }

    :host([striped]) .table tbody tr:nth-child(even) {
      background-color: var(--color-bg-subtle, rgb(249 250 251));
    }

    .table tbody td {
      padding: var(--space-4, 1rem) var(--space-6, 1.5rem);
      font-size: var(--font-size-sm, 0.875rem);
      color: var(--color-fg, rgb(17 24 39));
      vertical-align: top;
    }

    /* Size variants */
    :host([size="sm"]) .table thead th,
    :host([size="sm"]) .table tbody td {
      padding: var(--space-2, 0.5rem) var(--space-4, 1rem);
      font-size: var(--font-size-xs, 0.75rem);
    }

    /* Hover effects */
    .table tbody tr:hover {
      background-color: var(--color-bg-subtle, rgb(249 250 251));
    }

    /* Focus styles for keyboard navigation */
    .table tbody tr:focus-within {
      outline: 2px solid var(--color-info-bg, rgb(219 234 254));
      outline-offset: -2px;
    }

    /* State message styles */
    .state-message {
      padding: var(--space-4, 1rem);
      text-align: center;
      color: var(--color-fg-muted, rgb(107 114 128));
      font-size: var(--font-size-sm, 0.875rem);
    }

    .state-message.error {
      color: var(--color-danger-fg, rgb(220 38 38));
    }
  `;

  render() {
    // If any state is active, show state message instead of table content
    if (this.loading) {
      return html`
        <div class="table-container">
          <table class="table">
            <tbody>
              <tr>
                <td
                  colspan=${this.colSpan}
                  class="state-message"
                  data-testid="loading-message"
                >
                  ${this.loadingMessage}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      `;
    }

    if (this.errorMessage) {
      return html`
        <div class="table-container">
          <table class="table">
            <tbody>
              <tr>
                <td
                  colspan=${this.colSpan}
                  class="state-message error"
                  data-testid="error-message"
                >
                  ${this.errorMessage}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      `;
    }

    if (this.empty) {
      return html`
        <div class="table-container">
          <table class="table">
            <tbody>
              <tr>
                <td
                  colspan=${this.colSpan}
                  class="state-message"
                  data-testid="empty-message"
                >
                  ${this.emptyMessage}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      `;
    }

    // Default: show the slot content (thead, tbody)
    return html`
      <div class="table-container">
        <table class="table">
          <slot></slot>
        </table>
      </div>
    `;
  }
}

if (!customElements.get("ui-table")) {
  customElements.define("ui-table", UiTable);
}
