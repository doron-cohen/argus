import { LitElement, html } from "lit";
import { customElement } from "lit/decorators.js";
import "./router/outlet";

@customElement("argus-app")
export class ArgusApp extends LitElement {
  render() {
    return html`<router-outlet></router-outlet>`;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "argus-app": ArgusApp;
  }
}

// Initialize the app
const app = document.getElementById("app");
if (!app) {
  throw new Error("App container element not found");
}

// Use createElement instead of innerHTML for security
const appElement = document.createElement("argus-app");
app.appendChild(appElement);
