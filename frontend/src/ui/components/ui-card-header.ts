import { LitElement, html, css, type TemplateResult } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-card-header")
export class UiCardHeader extends LitElement {
  @property({ type: String })
  title = "";

  @property({ type: String })
  subtitle = "";

  @property({ type: String })
  titleDataTestId = "";

  @property({ type: String })
  subtitleDataTestId = "";

  static styles = css`
    :host {
      display: block;
      padding: var(--space-4, 1rem) var(--space-4, 1rem);
    }

    .title {
      margin: 0 0 var(--space-1, 0.25rem) 0;
      font-size: var(--font-size-lg, 1.125rem);
      font-weight: var(--font-weight-medium, 500);
      color: var(--color-fg, rgb(17 24 39));
      line-height: 1.5;
    }

    .subtitle {
      margin: 0;
      font-size: var(--font-size-sm, 0.875rem);
      color: var(--color-fg-muted, rgb(156 163 175));
      max-width: 42rem;
    }
  `;

  render(): TemplateResult {
    return html`
      ${this.title
        ? html`<h3
            class="title"
            data-testid="${this.titleDataTestId || "card-title"}"
          >
            ${this.title}
          </h3>`
        : ""}
      ${this.subtitle
        ? html`<p
            class="subtitle"
            data-testid="${this.subtitleDataTestId || "card-subtitle"}"
          >
            ${this.subtitle}
          </p>`
        : ""}
      <slot></slot>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "ui-card-header": UiCardHeader;
  }
}
