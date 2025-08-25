import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-page-container")
export class UiPageContainer extends LitElement {
  @property({ type: String, reflect: true })
  padding: "none" | "sm" | "md" | "lg" = "md";

  @property({ type: String, attribute: "max-width", reflect: true })
  maxWidth: "sm" | "md" | "lg" | "xl" | "full" = "lg";

  constructor() {
    super();
    this.padding = "md";
    this.maxWidth = "lg";
  }

  static styles = css`
    :host {
      display: block;
      width: 100%;
      margin: 0 auto;
    }

    :host([max-width="sm"]) {
      max-width: var(--max-width-sm, 640px);
    }

    :host([max-width="md"]) {
      max-width: var(--max-width-md, 768px);
    }

    :host([max-width="lg"]) {
      max-width: var(--max-width-lg, 1024px);
    }

    :host([max-width="xl"]) {
      max-width: var(--max-width-xl, 1280px);
    }

    :host([max-width="full"]) {
      max-width: none;
    }

    :host([padding="none"]) {
      padding: var(--space-0, 0);
    }

    :host([padding="sm"]) {
      padding: var(--space-4, 1rem);
    }

    :host([padding="md"]) {
      padding: var(--space-4, 1rem) var(--space-6, 1.5rem);
    }

    :host([padding="lg"]) {
      padding: var(--space-6, 1.5rem) var(--space-8, 2rem);
    }

    @media (min-width: 640px) {
      :host([padding="md"]) {
        padding: var(--space-6, 1.5rem) var(--space-8, 2rem);
      }

      :host([padding="lg"]) {
        padding: var(--space-8, 2rem) var(--space-12, 3rem);
      }
    }
  `;

  render() {
    return html`<slot></slot>`;
  }
}

if (!customElements.get("ui-page-container")) {
  customElements.define("ui-page-container", UiPageContainer);
}
