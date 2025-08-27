import { LitElement, html, css, type TemplateResult } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-info-row")
export class UiInfoRow extends LitElement {
  @property({ type: String })
  label = "";

  @property({ type: String })
  value = "";

  @property({ type: String })
  labelDataTestId = "";

  @property({ type: String })
  valueDataTestId = "";

  static styles = css`
    :host {
      display: block;
    }

    .info-row {
      background-color: var(--color-bg-subtle, rgb(249 250 251));
      padding: var(--space-4, 1rem);
      display: grid;
      grid-template-columns: 1fr;
      gap: var(--space-4, 1rem);
    }

    @media (min-width: 640px) {
      .info-row {
        grid-template-columns: 1fr 2fr;
        align-items: flex-start;
      }
    }

    .label {
      font-size: var(--font-size-sm, 0.875rem);
      font-weight: var(--font-weight-medium, 500);
      color: var(--color-fg-muted, rgb(156 163 175));
    }

    .value {
      font-size: var(--font-size-sm, 0.875rem);
      color: var(--color-fg, rgb(17 24 39));
      margin: 0;
    }

    @media (min-width: 640px) {
      .value {
        margin-top: 0;
      }
    }
  `;

  render(): TemplateResult {
    return html`
      <div class="info-row">
        <dt class="label" data-testid="${this.labelDataTestId || "info-label"}">
          ${this.label}
        </dt>
        <dd class="value" data-testid="${this.valueDataTestId || "info-value"}">
          ${this.value}
        </dd>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "ui-info-row": UiInfoRow;
  }
}
