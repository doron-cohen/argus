import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("ui-page-header")
export class UiPageHeader extends LitElement {
  @property({ type: String, reflect: true })
  title: string = "";

  @property({ type: String, reflect: true })
  description: string = "";

  @property({ type: String, reflect: true })
  size: "sm" | "md" | "lg" = "md";

  @property({ type: String })
  titleTestId = "page-title";

  @property({ type: String })
  descriptionTestId = "page-description";

  constructor() {
    super();
    this.title = "";
    this.description = "";
    this.size = "md";
  }

  static styles = css`
    :host {
      display: block;
      margin-bottom: var(--space-6, 1.5rem);
    }

    .header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: var(--space-4, 1rem);
    }

    .content {
      flex: 1;
      min-width: 0;
    }

    .title {
      margin: 0;
      color: var(--color-fg, rgb(17 24 39));
      font-weight: var(--font-weight-semibold, 600);
      line-height: 1.25;
    }

    .description {
      margin: var(--space-2, 0.5rem) 0 0 0;
      color: var(--color-fg-muted, rgb(107 114 128));
      font-size: var(--font-size-sm, 0.875rem);
      line-height: 1.5;
    }

    .actions {
      flex-shrink: 0;
      display: flex;
      align-items: center;
      gap: var(--space-2, 0.5rem);
    }

    /* Size variants */
    :host([size="sm"]) .title {
      font-size: var(--font-size-lg, 1.125rem);
    }

    :host([size="md"]) .title {
      font-size: var(--font-size-xl, 1.25rem);
    }

    :host([size="lg"]) .title {
      font-size: var(--font-size-2xl, 1.5rem);
    }

    @media (max-width: 640px) {
      .header {
        flex-direction: column;
        align-items: stretch;
        gap: var(--space-3, 0.75rem);
      }

      .actions {
        justify-content: flex-start;
      }
    }
  `;

  render() {
    return html`
      <div class="header">
        <div class="content">
          ${this.title
            ? html`<h1 class="title" data-testid="${this.titleTestId}">
                ${this.title}
              </h1>`
            : ""}
          ${this.description
            ? html`<p
                class="description"
                data-testid="${this.descriptionTestId}"
              >
                ${this.description}
              </p>`
            : ""}
        </div>
        <div class="actions">
          <slot name="actions"></slot>
        </div>
      </div>
    `;
  }
}

// Ensure element is registered even if decorators are not applied by the bundler
if (!customElements.get("ui-page-header")) {
  customElements.define("ui-page-header", UiPageHeader);
}
