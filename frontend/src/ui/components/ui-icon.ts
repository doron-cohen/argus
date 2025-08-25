import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

export type IconName = "check" | "x" | "warning" | "circle-check";

@customElement("ui-icon")
export class UiIcon extends LitElement {
  @property({ type: String, reflect: true })
  name: IconName = "check";

  @property({ type: String, reflect: true })
  size: "xs" | "sm" | "md" | "lg" = "sm";

  constructor() {
    super();
    this.name = "check";
    this.size = "sm";
  }

  static styles = css`
    :host {
      display: inline-block;
      width: var(--icon-size, 1rem);
      height: var(--icon-size, 1rem);
    }

    :host([size="xs"]) {
      --icon-size: 0.75rem;
    }

    :host([size="sm"]) {
      --icon-size: 1rem;
    }

    :host([size="md"]) {
      --icon-size: 1.25rem;
    }

    :host([size="lg"]) {
      --icon-size: 1.5rem;
    }

    svg {
      width: 100%;
      height: 100%;
      fill: currentColor;
    }
  `;

  render() {
    const path = this.getIconPath();
    if (!path) {
      console.warn(`Icon "${this.name}" not found`);
      return html``;
    }

    return html`
      <svg viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d=${path} clip-rule="evenodd"></path>
      </svg>
    `;
  }

  private getIconPath(): string {
    switch (this.name) {
      case "check":
        return "M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z";
      case "x":
        return "M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z";
      case "warning":
        return "M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z";
      case "circle-check":
        return "M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z";
      default:
        return "";
    }
  }
}

if (!customElements.get("ui-icon")) {
  customElements.define("ui-icon", UiIcon);
}
