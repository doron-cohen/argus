import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-card")
export class UiCard extends LitElement {
  @property({ type: String, reflect: true })
  variant: "default" | "elevated" | "outlined" = "default";

  @property({ type: String, reflect: true })
  padding: "none" | "sm" | "md" | "lg" = "md";

  constructor() {
    super();
    this.variant = "default";
    this.padding = "md";
  }

  static styles = css`
    :host {
      display: block;
      box-sizing: border-box;
      width: 100%;
      max-width: 100%;
      background-color: var(--color-bg, rgb(255 255 255));
      border-radius: var(--radius-2, 0.25rem);
      overflow: hidden;
    }

    /* Variant styles */
    :host([variant="default"]) {
      border: 1px solid var(--color-border, rgb(229 231 235));
      box-shadow: var(--shadow-1, 0 1px 2px 0 rgb(0 0 0 / 0.05));
    }

    :host([variant="elevated"]) {
      border: 1px solid var(--color-border, rgb(229 231 235));
      box-shadow: var(
        --shadow-2,
        0 1px 3px 0 rgb(0 0 0 / 0.1),
        0 1px 2px -1px rgb(0 0 0 / 0.1)
      );
    }

    :host([variant="outlined"]) {
      border: 1px solid var(--color-border, rgb(229 231 235));
      box-shadow: none;
    }

    /* Padding variants */
    :host([padding="none"]) .content {
      padding: var(--space-0, 0);
    }

    :host([padding="sm"]) .content {
      padding: var(--space-3, 0.75rem);
    }

    :host([padding="md"]) .content {
      padding: var(--space-4, 1rem);
    }

    :host([padding="lg"]) .content {
      padding: var(--space-6, 1.5rem);
    }

    .header {
      padding: var(--space-4, 1rem);
      border-bottom: 1px solid var(--color-border, rgb(229 231 235));
      background-color: var(--color-bg-subtle, rgb(249 250 251));
    }

    .header:empty {
      display: none;
    }

    .content {
      padding: var(--space-4, 1rem);
    }

    .footer {
      padding: var(--space-4, 1rem);
      border-top: 1px solid var(--color-border, rgb(229 231 235));
      background-color: var(--color-bg-subtle, rgb(249 250 251));
    }

    .footer:empty {
      display: none;
    }

    /* Header and footer padding adjustments when content has different padding */
    :host([padding="sm"]) .header,
    :host([padding="sm"]) .footer {
      padding: var(--space-3, 0.75rem);
    }

    :host([padding="lg"]) .header,
    :host([padding="lg"]) .footer {
      padding: var(--space-6, 1.5rem);
    }

    :host([padding="none"]) .header,
    :host([padding="none"]) .footer {
      padding: var(--space-4, 1rem);
    }
  `;

  render() {
    return html`
      <div class="header">
        <slot name="header"></slot>
      </div>
      <div class="content">
        <slot></slot>
      </div>
      <div class="footer">
        <slot name="footer"></slot>
      </div>
    `;
  }
}

if (!customElements.get("ui-card")) {
  customElements.define("ui-card", UiCard);
}
