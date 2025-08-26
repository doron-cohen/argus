import { LitElement, html, nothing, type TemplateResult } from "lit";
import { customElement, property } from "lit/decorators.js";
import "../tokens/core.css";
import "../tokens/semantic.css";

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
      <div class="text-center py-8">
        ${this.icon
          ? html`<ui-icon name=${this.icon} class="mx-auto h-12 w-12 text-gray-400 mb-4"></ui-icon>`
          : nothing}
        <div class="u-text-muted text-lg">${this.title}</div>
        ${this.description
          ? html`<div class="u-text-muted text-sm mt-2">${this.description}</div>`
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
