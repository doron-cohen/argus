import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-stack")
export class UiStack extends LitElement {
  @property({ type: String, reflect: true })
  gap: "xs" | "sm" | "md" | "lg" | "xl" | "2xl" = "md";

  @property({ type: Boolean, reflect: true })
  splitAfter: number | null = null;

  constructor() {
    super();
    this.gap = "md";
    this.splitAfter = null;
  }

  static styles = css`
    :host {
      display: flex;
      flex-direction: column;
      gap: var(--stack-gap, var(--space-4, 1rem));
    }

    :host([gap="xs"]) {
      --stack-gap: var(--space-2, 0.5rem);
    }

    :host([gap="sm"]) {
      --stack-gap: var(--space-3, 0.75rem);
    }

    :host([gap="md"]) {
      --stack-gap: var(--space-4, 1rem);
    }

    :host([gap="lg"]) {
      --stack-gap: var(--space-6, 1.5rem);
    }

    :host([gap="xl"]) {
      --stack-gap: var(--space-8, 2rem);
    }

    :host([gap="2xl"]) {
      --stack-gap: var(--space-12, 3rem);
    }

    ::slotted(*) {
      margin: 0;
    }
  `;

  render() {
    return html`<slot></slot>`;
  }
}

if (!customElements.get("ui-stack")) {
  customElements.define("ui-stack", UiStack);
}
