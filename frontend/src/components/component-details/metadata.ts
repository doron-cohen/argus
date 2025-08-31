import { LitElement, html } from "lit";
import { customElement, property } from "lit/decorators.js";
import type { Component } from "../../api/services/components/client";
import "../../ui/components/ui-card.js";
import "../../ui/components/ui-card-header.js";
import "../../ui/components/ui-info-row.js";

@customElement("component-metadata")
export class ComponentMetadata extends LitElement {
  @property({ type: Object, attribute: false })
  component: Component | null = null;

  render() {
    if (!this.component) {
      return html`<div>No component data available</div>`;
    }

    // Since component is validated as non-null, extract to local variable
    const component = this.component;

    return html`
      ${this.renderComponentHeader(component)}
      ${this.renderComponentInfo(component)}
    `;
  }

  private renderComponentHeader(component: Component) {
    return html`
      <ui-card>
        <ui-card-header
          title="${component.name}"
          subtitle="ID: ${component.id ?? component.name}"
          title-data-testid="component-name"
          subtitle-data-testid="component-id"
        ></ui-card-header>
      </ui-card>
    `;
  }

  private renderComponentInfo(component: Component) {
    return html`
      <ui-card class="u-mt-4">
        ${this.renderInfoRow(
          "Description",
          "description-label",
          "component-description",
          component.description || "No description available",
        )}
        ${this.renderInfoRow(
          "Team",
          "team-label",
          "component-team",
          component.owners?.team || "No team assigned",
        )}
        ${this.renderMaintainersRow(component)}
      </ui-card>
    `;
  }

  private renderInfoRow(
    label: string,
    labelTestId: string,
    valueTestId: string,
    value: string,
  ) {
    return html`
      <ui-info-row
        label="${label}"
        value="${value}"
        label-data-testid="${labelTestId}"
        value-data-testid="${valueTestId}"
      ></ui-info-row>
    `;
  }

  private renderMaintainersRow(component: Component) {
    const maintainers = component.owners?.maintainers;
    if (!maintainers?.length) {
      return this.renderInfoRow(
        "Maintainers",
        "maintainers-label",
        "component-maintainers",
        "No maintainers assigned",
      );
    }

    return html`
      <ui-info-row
        label="Maintainers"
        value="${maintainers.join(", ")}"
        label-data-testid="maintainers-label"
        value-data-testid="component-maintainers"
      ></ui-info-row>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "component-metadata": ComponentMetadata;
  }
}
