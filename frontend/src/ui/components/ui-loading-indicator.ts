import { LitElement, html, nothing, type TemplateResult } from "lit";
import { customElement, property } from "lit/decorators.js";
import "../tokens/core.css";
import "../tokens/semantic.css";

@customElement("ui-loading-indicator")
export class UiLoadingIndicator extends LitElement {
  @property({ type: String })
  message = "Loading...";

  @property({ type: String })
  size = "md"; // sm, md, lg

  render(): TemplateResult {
    const spinnerSize =
      this.size === "sm"
        ? "h-4 w-4"
        : this.size === "lg"
          ? "h-8 w-8"
          : "h-6 w-6";

    return html`
      <div class="flex items-center justify-center space-x-2 py-8">
        <div
          class="animate-spin rounded-full border-b-2 border-blue-600 ${spinnerSize}"
        ></div>
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
