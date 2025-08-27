import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-grid")
export class UiGrid extends LitElement {
  @property({ type: String, reflect: true })
  cols: "1" | "2" | "3" | "4" | "auto" = "1";

  @property({ type: String, reflect: true })
  gap: "xs" | "sm" | "md" | "lg" | "xl" = "md";

  @property({ type: String, reflect: true })
  breakpoint: "sm" | "md" | "lg" | "xl" = "lg";

  constructor() {
    super();
    this.cols = "1";
    this.gap = "md";
    this.breakpoint = "lg";
  }

  static styles = css`
    :host {
      display: grid;
      grid-template-columns: 1fr;
      gap: var(--grid-gap, var(--space-4, 1rem));
      width: 100%;
    }

    :host([cols="1"]) {
      grid-template-columns: 1fr;
    }

    :host([cols="2"]) {
      grid-template-columns: repeat(2, 1fr);
    }

    :host([cols="3"]) {
      grid-template-columns: repeat(3, 1fr);
    }

    :host([cols="4"]) {
      grid-template-columns: repeat(4, 1fr);
    }

    :host([cols="auto"]) {
      grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    }

    /* Breakpoint variants */
    @media (min-width: 640px) {
      :host([breakpoint="sm"][cols="2"]) {
        grid-template-columns: repeat(2, 1fr);
      }

      :host([breakpoint="sm"][cols="3"]) {
        grid-template-columns: repeat(3, 1fr);
      }

      :host([breakpoint="sm"][cols="4"]) {
        grid-template-columns: repeat(4, 1fr);
      }
    }

    @media (min-width: 768px) {
      :host([breakpoint="md"][cols="2"]) {
        grid-template-columns: repeat(2, 1fr);
      }

      :host([breakpoint="md"][cols="3"]) {
        grid-template-columns: repeat(3, 1fr);
      }

      :host([breakpoint="md"][cols="4"]) {
        grid-template-columns: repeat(4, 1fr);
      }
    }

    @media (min-width: 1024px) {
      :host([breakpoint="lg"][cols="2"]) {
        grid-template-columns: repeat(2, 1fr);
      }

      :host([breakpoint="lg"][cols="3"]) {
        grid-template-columns: repeat(3, 1fr);
      }

      :host([breakpoint="lg"][cols="4"]) {
        grid-template-columns: repeat(4, 1fr);
      }
    }

    @media (min-width: 1280px) {
      :host([breakpoint="xl"][cols="2"]) {
        grid-template-columns: repeat(2, 1fr);
      }

      :host([breakpoint="xl"][cols="3"]) {
        grid-template-columns: repeat(3, 1fr);
      }

      :host([breakpoint="xl"][cols="4"]) {
        grid-template-columns: repeat(4, 1fr);
      }
    }

    /* Gap variants */
    :host([gap="xs"]) {
      --grid-gap: var(--space-2, 0.5rem);
    }

    :host([gap="sm"]) {
      --grid-gap: var(--space-3, 0.75rem);
    }

    :host([gap="md"]) {
      --grid-gap: var(--space-4, 1rem);
    }

    :host([gap="lg"]) {
      --grid-gap: var(--space-6, 1.5rem);
    }

    :host([gap="xl"]) {
      --grid-gap: var(--space-8, 2rem);
    }

    ::slotted(*) {
      margin: 0;
    }
  `;

  render() {
    return html`<slot></slot>`;
  }
}

if (!customElements.get("ui-grid")) {
  customElements.define("ui-grid", UiGrid);
}
