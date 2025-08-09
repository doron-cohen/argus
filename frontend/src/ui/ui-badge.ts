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

export class UiBadge extends HTMLElement {
  static get observedAttributes() {
    return ["status"] as const;
  }

  connectedCallback(): void {
    this.render();
  }

  attributeChangedCallback(): void {
    this.render();
  }

  private render(): void {
    const status = normalizeStatus(this.getAttribute("status"));
    const colorClasses = getColorClasses(status);

    // Keep the classes on the host element for test compatibility and simplicity
    this.className = `${colorClasses} px-2 py-1 rounded-full text-xs font-medium flex items-center`;

    const icon = getIconSvg(status);
    const label = escapeHtml(this.getAttribute("status") || status);

    this.innerHTML = `${icon}${label}`;
  }
}

if (!customElements.get("ui-badge")) {
  customElements.define("ui-badge", UiBadge);
}
