import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-box")
export class UiBox extends LitElement {
  @property({ type: String, reflect: true })
  display:
    | "block"
    | "inline"
    | "inline-block"
    | "flex"
    | "inline-flex"
    | "grid"
    | "inline-grid" = "block";

  @property({ type: String, reflect: true })
  padding: "none" | "xs" | "sm" | "md" | "lg" | "xl" = "none";

  @property({ type: String, reflect: true })
  margin: "none" | "xs" | "sm" | "md" | "lg" | "xl" = "none";

  @property({ type: String, reflect: true })
  border: "none" | "thin" | "thick" = "none";

  @property({ type: String, reflect: true })
  radius: "none" | "sm" | "md" | "lg" | "full" = "none";

  @property({ type: String, reflect: true })
  background: "none" | "subtle" | "muted" | "primary" | "secondary" = "none";

  constructor() {
    super();
    this.display = "block";
    this.padding = "none";
    this.margin = "none";
    this.border = "none";
    this.radius = "none";
    this.background = "none";
  }

  static styles = css`
    :host {
      display: var(--ui-box-display, block);
      box-sizing: border-box;
      width: 100%;
    }

    /* Display variants */
    :host([display="block"]) {
      --ui-box-display: block;
    }

    :host([display="inline"]) {
      --ui-box-display: inline;
      width: auto;
    }

    :host([display="inline-block"]) {
      --ui-box-display: inline-block;
      width: auto;
    }

    :host([display="flex"]) {
      --ui-box-display: flex;
    }

    :host([display="inline-flex"]) {
      --ui-box-display: inline-flex;
      width: auto;
    }

    :host([display="grid"]) {
      --ui-box-display: grid;
    }

    :host([display="inline-grid"]) {
      --ui-box-display: inline-grid;
      width: auto;
    }

    /* Padding variants */
    :host([padding="none"]) {
      padding: var(--space-0, 0);
    }

    :host([padding="xs"]) {
      padding: var(--space-1, 0.25rem);
    }

    :host([padding="sm"]) {
      padding: var(--space-2, 0.5rem);
    }

    :host([padding="md"]) {
      padding: var(--space-4, 1rem);
    }

    :host([padding="lg"]) {
      padding: var(--space-6, 1.5rem);
    }

    :host([padding="xl"]) {
      padding: var(--space-8, 2rem);
    }

    /* Margin variants */
    :host([margin="none"]) {
      margin: var(--space-0, 0);
    }

    :host([margin="xs"]) {
      margin: var(--space-1, 0.25rem);
    }

    :host([margin="sm"]) {
      margin: var(--space-2, 0.5rem);
    }

    :host([margin="md"]) {
      margin: var(--space-4, 1rem);
    }

    :host([margin="lg"]) {
      margin: var(--space-6, 1.5rem);
    }

    :host([margin="xl"]) {
      margin: var(--space-8, 2rem);
    }

    /* Border variants */
    :host([border="none"]) {
      border: none;
    }

    :host([border="thin"]) {
      border: 1px solid var(--color-border, rgb(229 231 235));
    }

    :host([border="thick"]) {
      border: 2px solid var(--color-border, rgb(229 231 235));
    }

    /* Radius variants */
    :host([radius="none"]) {
      border-radius: var(--radius-0, 0);
    }

    :host([radius="sm"]) {
      border-radius: var(--radius-2, 0.25rem);
    }

    :host([radius="md"]) {
      border-radius: var(--radius-4, 0.5rem);
    }

    :host([radius="lg"]) {
      border-radius: var(--radius-8, 0.75rem);
    }

    :host([radius="full"]) {
      border-radius: var(--radius-full, 9999px);
    }

    /* Background variants */
    :host([background="none"]) {
      background-color: transparent;
    }

    :host([background="subtle"]) {
      background-color: var(--color-bg-subtle, rgb(249 250 251));
    }

    :host([background="muted"]) {
      background-color: var(--color-bg-muted, rgb(243 244 246));
    }

    :host([background="primary"]) {
      background-color: var(--color-primary-bg, rgb(59 130 246));
      color: var(--color-primary-fg, rgb(255 255 255));
    }

    :host([background="secondary"]) {
      background-color: var(--color-secondary-bg, rgb(107 114 128));
      color: var(--color-secondary-fg, rgb(255 255 255));
    }
  `;

  render() {
    return html`<slot></slot>`;
  }
}

if (!customElements.get("ui-box")) {
  customElements.define("ui-box", UiBox);
}
