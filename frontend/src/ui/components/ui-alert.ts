import { LitElement, html, css, type TemplateResult } from "lit";
import { customElement, property } from "lit/decorators.js";

export type AlertVariant = "success" | "error" | "warning" | "info";

@customElement("ui-alert")
export class UiAlert extends LitElement {
  @property({ type: String, reflect: true })
  variant: AlertVariant = "info";

  @property({ type: String })
  title = "";

  @property({ type: String })
  message = "";

  @property({ type: Boolean })
  dismissible = false;

  static styles = css`
    :host {
      display: block;
      border-radius: var(--radius-2, 0.25rem);
      border: 1px solid;
      padding: var(--space-4, 1rem);
      position: relative;
    }

    .alert-content {
      display: flex;
      align-items: flex-start;
      gap: var(--space-3, 0.75rem);
    }

    .alert-icon {
      flex-shrink: 0;
      width: 1.25rem;
      height: 1.25rem;
      margin-top: 0.125rem;
    }

    .alert-text {
      flex: 1;
    }

    .alert-title {
      font-weight: var(--font-weight-semibold, 600);
      margin-bottom: var(--space-1, 0.25rem);
      font-size: var(--font-size-sm, 0.875rem);
    }

    .alert-message {
      font-size: var(--font-size-sm, 0.875rem);
      margin: 0;
    }

    .dismiss-button {
      background: none;
      border: none;
      cursor: pointer;
      padding: 0.25rem;
      margin-left: var(--space-2, 0.5rem);
      border-radius: var(--radius-1, 0.125rem);
      color: inherit;
      opacity: 0.7;
      transition: opacity 0.15s ease;
    }

    .dismiss-button:hover {
      opacity: 1;
    }

    .dismiss-button:focus {
      outline: 2px solid currentColor;
      outline-offset: 2px;
    }

    /* Variant styles */
    :host([variant="success"]) {
      background-color: var(--color-success-bg, rgb(236 253 245));
      color: var(--color-success-fg, rgb(34 197 94));
      border-color: var(--color-success-border, rgb(187 247 208));
    }

    :host([variant="error"]) {
      background-color: var(--color-danger-bg, rgb(254 242 242));
      color: var(--color-danger-fg, rgb(239 68 68));
      border-color: var(--color-danger-border, rgb(252 165 165));
    }

    :host([variant="warning"]) {
      background-color: var(--color-warning-bg, rgb(255 251 235));
      color: var(--color-warning-fg, rgb(245 158 11));
      border-color: var(--color-warning-border, rgb(254 240 138));
    }

    :host([variant="info"]) {
      background-color: var(--color-info-bg, rgb(239 246 255));
      color: var(--color-info-fg, rgb(59 130 246));
      border-color: var(--color-info-border, rgb(191 219 254));
    }
  `;

  private getIconName(): string {
    switch (this.variant) {
      case "success":
        return "check-circle";
      case "error":
        return "x-circle";
      case "warning":
        return "exclamation-triangle";
      case "info":
      default:
        return "information-circle";
    }
  }

  private handleDismiss(): void {
    this.dispatchEvent(new CustomEvent("dismiss"));
    this.remove();
  }

  render(): TemplateResult {
    return html`
      <div class="alert-content">
        <ui-icon name="${this.getIconName()}" class="alert-icon"></ui-icon>
        <div class="alert-text">
          ${this.title
            ? html`<div class="alert-title">${this.title}</div>`
            : ""}
          ${this.message
            ? html`<p class="alert-message">${this.message}</p>`
            : ""}
          <slot></slot>
        </div>
        ${this.dismissible
          ? html`
              <button
                class="dismiss-button"
                @click=${this.handleDismiss}
                aria-label="Dismiss alert"
              >
                <ui-icon name="x"></ui-icon>
              </button>
            `
          : ""}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "ui-alert": UiAlert;
  }
}
