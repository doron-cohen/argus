import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-cluster")
export class UiCluster extends LitElement {
  @property({ type: String, reflect: true })
  justify: "start" | "center" | "end" | "between" | "around" | "evenly" =
    "start";

  @property({ type: String, reflect: true })
  align: "start" | "center" | "end" | "baseline" | "stretch" = "center";

  @property({ type: String, reflect: true })
  gap: "xs" | "sm" | "md" | "lg" | "xl" = "md";

  constructor() {
    super();
    this.justify = "start";
    this.align = "center";
    this.gap = "md";
  }

  static styles = css`
    :host {
      display: flex;
      flex-direction: row;
      align-items: var(--cluster-align, center);
      justify-content: var(--cluster-justify, flex-start);
      gap: var(--cluster-gap, var(--space-4, 1rem));
      width: 100%;
    }

    :host([justify="start"]) {
      --cluster-justify: flex-start;
    }

    :host([justify="center"]) {
      --cluster-justify: center;
    }

    :host([justify="end"]) {
      --cluster-justify: flex-end;
    }

    :host([justify="between"]) {
      --cluster-justify: space-between;
    }

    :host([justify="around"]) {
      --cluster-justify: space-around;
    }

    :host([justify="evenly"]) {
      --cluster-justify: space-evenly;
    }

    :host([align="start"]) {
      --cluster-align: flex-start;
    }

    :host([align="center"]) {
      --cluster-align: center;
    }

    :host([align="end"]) {
      --cluster-align: flex-end;
    }

    :host([align="baseline"]) {
      --cluster-align: baseline;
    }

    :host([align="stretch"]) {
      --cluster-align: stretch;
    }

    :host([gap="xs"]) {
      --cluster-gap: var(--space-2, 0.5rem);
    }

    :host([gap="sm"]) {
      --cluster-gap: var(--space-3, 0.75rem);
    }

    :host([gap="md"]) {
      --cluster-gap: var(--space-4, 1rem);
    }

    :host([gap="lg"]) {
      --cluster-gap: var(--space-6, 1.5rem);
    }

    :host([gap="xl"]) {
      --cluster-gap: var(--space-8, 2rem);
    }

    ::slotted(*) {
      margin: 0;
      flex-shrink: 0;
    }

    ::slotted(*:last-child) {
      margin-left: auto;
    }
  `;

  render() {
    return html`<slot></slot>`;
  }
}

if (!customElements.get("ui-cluster")) {
  customElements.define("ui-cluster", UiCluster);
}
