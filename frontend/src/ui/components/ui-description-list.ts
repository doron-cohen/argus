import { LitElement, html, type TemplateResult } from "lit";
import { customElement, property } from "lit/decorators.js";

export interface DescriptionItem {
  label: string;
  value: string;
}

@customElement("ui-description-list")
export class UiDescriptionList extends LitElement {
  @property({ type: Array })
  items: DescriptionItem[] = [];

  render(): TemplateResult {
    if (!this.items.length) {
      return html``;
    }

    return html`
      <div class="u-stack-2">
        ${this.items.map(
          (item) => html`
            <div>
              <span class="u-font-medium u-text-secondary">${item.label}:</span>
              <span class="u-text-primary">${item.value || "N/A"}</span>
            </div>
          `
        )}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "ui-description-list": UiDescriptionList;
  }
}
