import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";
import { escapeHtml } from "../utils";

type StatusVariant =
  | "pass"
  | "fail"
  | "error"
  | "unknown"
  | "disabled"
  | "skipped"
  | "completed"
  | "default";

function getColorClasses(status: StatusVariant): string {
  switch (status) {
    case "pass":
      return "bg-green-100 text-green-800";
    case "fail":
    case "error":
    case "unknown":
      return "bg-red-100 text-red-800";
    case "disabled":
    case "skipped":
      return "bg-yellow-100 text-yellow-800";
    case "completed":
      return "bg-blue-100 text-blue-800";
    default:
      return "bg-gray-100 text-gray-800";
  }
}

function getIconSvg(status: StatusVariant): string {
  switch (status) {
    case "pass":
      return '<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"></path></svg>';
    case "fail":
    case "error":
    case "unknown":
      return '<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path></svg>';
    case "disabled":
    case "skipped":
      return '<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path></svg>';
    case "completed":
      return '<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path></svg>';
    default:
      return "";
  }
}

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
      padding: 0.25rem 0.5rem;
      border-radius: 9999px;
      font-size: 0.75rem;
      font-weight: 500;
    }

    .icon {
      width: 0.75rem;
      height: 0.75rem;
      margin-right: 0.25rem;
    }

    /* Status-based styling */
    :host(.pass) {
      background-color: rgb(220 252 231);
      color: rgb(22 163 74);
    }

    :host(.fail),
    :host(.error),
    :host(.unknown) {
      background-color: rgb(254 226 226);
      color: rgb(220 38 38);
    }

    :host(.disabled),
    :host(.skipped) {
      background-color: rgb(254 249 195);
      color: rgb(161 98 7);
    }

    :host(.completed) {
      background-color: rgb(219 234 254);
      color: rgb(29 78 216);
    }

    :host(.default) {
      background-color: rgb(243 244 246);
      color: rgb(55 65 81);
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
    const icon = getIconSvg(status);
    const label = escapeHtml(this.status || status);

    return html`
      ${icon ? html`<span class="icon" .innerHTML=${icon}></span>` : ""}
      ${label}
    `;
  }
}
