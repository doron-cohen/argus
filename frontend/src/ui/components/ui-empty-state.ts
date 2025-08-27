import { LitElement, html, nothing, type TemplateResult } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-empty-state")
export class UiEmptyState extends LitElement {
  @property({ type: String })
  title = "No data available";

  @property({ type: String })
  description = "";

  @property({ type: String })
  icon = "";

  render(): TemplateResult {
    return html`
      <div class="u-text-center u-py-8">
        ${this.icon
          ? html`<ui-icon
              name=${this.icon}
              class="u-block u-mx-auto u-h-12 u-w-12 u-text-muted u-mb-4"
            ></ui-icon>`
          : nothing}
        <div class="u-text-muted u-text-lg">${this.title}</div>
        ${this.description
          ? html`<div class="u-text-muted u-text-sm u-mt-2">
              ${this.description}
            </div>`
          : nothing}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "ui-empty-state": UiEmptyState;
  }
}
