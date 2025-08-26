import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";
import { escapeHtml } from "../../utils.js";
import "./ui-icon.js";

type StatusVariant =
  | "pass"
  | "fail"
  | "error"
  | "unknown"
  | "disabled"
  | "skipped"
  | "completed"
  | "default";

function normalizeStatus(value: string | null): StatusVariant {
  const v = (value || "").toLowerCase();
  switch (v) {
    case "pass":
    case "fail":
    case "error":
    case "unknown":
    case "disabled":
    case "skipped":
    case "completed":
      return v as StatusVariant;
    default:
      return "default";
  }
}

@customElement("ui-badge")
export class UiBadge extends LitElement {
  @property({ type: String, attribute: true })
  status: string = "default";

  static styles = css`
    :host {
      display: inline-flex;
      align-items: center;
      padding: var(--space-1, 0.25rem) var(--space-2, 0.5rem);
      border-radius: var(--radius-pill, 9999px);
      font-size: var(--font-size-xs, 0.75rem);
      font-weight: var(--font-weight-medium, 500);
      background-color: var(--color-neutral-bg, rgb(243 244 246));
      color: var(--color-neutral-fg, rgb(55 65 81));
    }

    /* CSS Parts for customization */
    .badge {
      display: flex;
      align-items: center;
      gap: var(--space-1, 0.25rem);
    }

    .icon {
      width: 0.75rem;
      height: 0.75rem;
      flex-shrink: 0;
    }

    .label {
      flex: 1;
    }

    /* Status-based styling */
    :host(.pass) {
      background-color: var(--color-success-bg, rgb(220 252 231));
      color: var(--color-success-fg, rgb(22 163 74));
    }

    :host(.fail),
    :host(.error),
    :host(.unknown) {
      background-color: var(--color-danger-bg, rgb(254 226 226));
      color: var(--color-danger-fg, rgb(220 38 38));
    }

    :host(.disabled),
    :host(.skipped) {
      background-color: var(--color-warning-bg, rgb(254 249 195));
      color: var(--color-warning-fg, rgb(161 98 7));
    }

    :host(.completed) {
      background-color: var(--color-info-bg, rgb(219 234 254));
      color: var(--color-info-fg, rgb(29 78 216));
    }

    :host(.default) {
      background-color: var(--color-neutral-bg, rgb(243 244 246));
      color: var(--color-neutral-fg, rgb(55 65 81));
    }
  `;

  updated(changedProperties: Map<string, any>) {
    super.updated(changedProperties);
    if (changedProperties.has("status")) {
      this.updateHostClasses();
    }
  }

  connectedCallback() {
    super.connectedCallback();
    this.updateHostClasses();
  }

  private updateHostClasses() {
    const status = normalizeStatus(this.status);

    // Remove all status classes
    this.classList.remove(
      "pass",
      "fail",
      "error",
      "unknown",
      "disabled",
      "skipped",
      "completed",
      "default",
    );

    // Add the current status class
    this.classList.add(status);
  }

  render() {
    const status = normalizeStatus(this.status);
    const label = escapeHtml(this.status || status);

    return html`
      <div class="badge" part="badge">
        ${this.renderIcon(status)}
        <span class="label" part="label">${label}</span>
      </div>
    `;
  }

  private renderIcon(status: StatusVariant) {
    switch (status) {
      case "pass":
        return html`<ui-icon class="icon" name="check" size="xs" part="icon"></ui-icon>`;
      case "fail":
      case "error":
      case "unknown":
        return html`<ui-icon class="icon" name="x" size="xs" part="icon"></ui-icon>`;
      case "disabled":
      case "skipped":
        return html`<ui-icon class="icon" name="warning" size="xs" part="icon"></ui-icon>`;
      case "completed":
        return html`<ui-icon
          class="icon"
          name="circle-check"
          size="xs"
          part="icon"
        ></ui-icon>`;
      default:
        return "";
    }
  }
}
