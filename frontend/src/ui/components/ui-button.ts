import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

type ButtonVariant = "primary" | "secondary" | "ghost";
type ButtonSize = "sm" | "md" | "lg";

@customElement("ui-button")
export class UiButton extends LitElement {
  @property({ type: String, reflect: true })
  variant: ButtonVariant = "primary";

  @property({ type: String, reflect: true })
  size: ButtonSize = "md";

  @property({ type: Boolean, reflect: true })
  disabled = false;

  static styles = css`
    :host {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      font-weight: var(--font-weight-medium, 500);
      border-radius: var(--radius-2, 0.25rem);
      border: 1px solid transparent;
      cursor: pointer;
      transition: all 0.15s ease;
      text-decoration: none;
      white-space: nowrap;
    }

    :host([disabled]) {
      cursor: not-allowed;
      opacity: 0.5;
    }

    button {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: var(--space-2, 0.5rem);
      font: inherit;
      background: none;
      border: none;
      cursor: inherit;
      width: 100%;
    }

    /* Primary variant */
    :host([variant="primary"]) {
      background-color: var(--color-info-bg, rgb(219 234 254));
      color: var(--color-info-fg, rgb(29 78 216));
      border-color: var(--color-info-bg, rgb(219 234 254));
    }

    :host([variant="primary"]:not([disabled]):hover) {
      background-color: var(--color-info-hover, rgb(191 219 254));
      border-color: var(--color-info-hover, rgb(191 219 254));
    }

    :host([variant="primary"]:not([disabled]):active) {
      background-color: var(--color-info-bg, rgb(219 234 254));
      transform: scale(0.98);
    }

    /* Secondary variant */
    :host([variant="secondary"]) {
      background-color: var(--color-neutral-bg, rgb(243 244 246));
      color: var(--color-neutral-fg, rgb(55 65 81));
      border-color: var(--color-border, rgb(229 231 235));
    }

    :host([variant="secondary"]:not([disabled]):hover) {
      background-color: var(--color-bg-subtle, rgb(249 250 251));
    }

    :host([variant="secondary"]:not([disabled]):active) {
      background-color: var(--color-neutral-bg, rgb(243 244 246));
      transform: scale(0.98);
    }

    /* Ghost variant */
    :host([variant="ghost"]) {
      background-color: transparent;
      color: var(--color-fg, rgb(17 24 39));
      border-color: transparent;
    }

    :host([variant="ghost"]:not([disabled]):hover) {
      background-color: var(--color-bg-subtle, rgb(249 250 251));
    }

    :host([variant="ghost"]:not([disabled]):active) {
      background-color: var(--color-neutral-bg, rgb(243 244 246));
      transform: scale(0.98);
    }

    /* Size variants */
    :host([size="sm"]) {
      font-size: var(--font-size-xs, 0.75rem);
      padding: var(--space-1-5, 0.375rem) var(--space-3, 0.75rem);
      min-height: var(--space-7, 1.75rem);
    }

    :host([size="md"]) {
      font-size: var(--font-size-sm, 0.875rem);
      padding: var(--space-2, 0.5rem) var(--space-4, 1rem);
      min-height: var(--space-9, 2.25rem);
    }

    :host([size="lg"]) {
      font-size: var(--font-size-md, 1rem);
      padding: var(--space-2-5, 0.625rem) var(--space-5, 1.25rem);
      min-height: var(--space-11, 2.75rem);
    }

    /* Icon-only buttons */
    :host(.icon-only) button {
      padding: 0;
      width: var(--space-9, 2.25rem);
      height: var(--space-9, 2.25rem);
    }

    :host(.icon-only[size="sm"]) button {
      width: var(--space-7, 1.75rem);
      height: var(--space-7, 1.75rem);
    }

    :host(.icon-only[size="lg"]) button {
      width: var(--space-11, 2.75rem);
      height: var(--space-11, 2.75rem);
    }

    /* CSS Parts for customization */
    button::part(label) {
      flex: 1;
    }

    button::part(icon) {
      flex-shrink: 0;
    }
  `;

  render() {
    return html`
      <button ?disabled=${this.disabled} part="button">
        <slot name="icon-start" part="icon"></slot>
        <slot part="label"></slot>
        <slot name="icon-end" part="icon"></slot>
      </button>
    `;
  }
}

if (!customElements.get("ui-button")) {
  customElements.define("ui-button", UiButton);
}
