import { LitElement, html, css, type TemplateResult } from "lit";
import { customElement, property } from "lit/decorators.js";

export type SpinnerSize = "xs" | "sm" | "md" | "lg" | "xl";

@customElement("ui-spinner")
export class UiSpinner extends LitElement {
  @property({ type: String, reflect: true })
  size: SpinnerSize = "md";

  @property({ type: String, reflect: true })
  color: "primary" | "secondary" | "muted" = "primary";

  static styles = css`
    :host {
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }

    .spinner {
      border: 2px solid transparent;
      border-top-color: currentColor;
      border-radius: 50%;
      animation: spin 1s linear infinite;
    }

    /* Size variants */
    :host([size="xs"]) .spinner {
      width: 0.75rem;
      height: 0.75rem;
    }

    :host([size="sm"]) .spinner {
      width: 1rem;
      height: 1rem;
    }

    :host([size="md"]) .spinner {
      width: 1.5rem;
      height: 1.5rem;
    }

    :host([size="lg"]) .spinner {
      width: 2rem;
      height: 2rem;
    }

    :host([size="xl"]) .spinner {
      width: 3rem;
      height: 3rem;
    }

    /* Color variants */
    :host([color="primary"]) {
      color: var(--color-info-fg, rgb(59 130 246));
    }

    :host([color="secondary"]) {
      color: var(--color-secondary-fg, rgb(107 114 128));
    }

    :host([color="muted"]) {
      color: var(--color-fg-muted, rgb(156 163 175));
    }

    @keyframes spin {
      from {
        transform: rotate(0deg);
      }
      to {
        transform: rotate(360deg);
      }
    }
  `;

  render(): TemplateResult {
    return html`<div class="spinner"></div>`;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "ui-spinner": UiSpinner;
  }
}
