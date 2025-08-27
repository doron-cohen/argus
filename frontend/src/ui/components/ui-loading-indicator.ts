import { LitElement, html, nothing, type TemplateResult } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-loading-indicator")
export class UiLoadingIndicator extends LitElement {
  @property({ type: String })
  message = "Loading...";

  @property({ type: String })
  size = "md"; // sm, md, lg

  render(): TemplateResult {
    const spinnerSize =
      this.size === "sm" ? "xs" : this.size === "lg" ? "lg" : "md";

    return html`
      <div class="u-flex u-items-center u-justify-center u-gap-2 u-py-8">
        <ui-spinner size="${spinnerSize}" color="primary"></ui-spinner>
        <span class="u-text-muted">${this.message}</span>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "ui-loading-indicator": UiLoadingIndicator;
  }
}
