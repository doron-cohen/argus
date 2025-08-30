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

    return html`
      ${this.renderComponentHeader()} ${this.renderComponentInfo()}
    `;
  }

  private renderComponentHeader() {
    return html`
      <ui-card>
        <ui-card-header
          title="${this.component!.name}"
          subtitle="ID: ${this.component!.id || this.component!.name}"
          title-data-testid="component-name"
          subtitle-data-testid="component-id"
        ></ui-card-header>
      </ui-card>
    `;
  }

  private renderComponentInfo() {
    return html`
      <ui-card class="u-mt-4">
        ${this.renderInfoRow(
          "Description",
          "description-label",
          "component-description",
          this.component!.description || "No description available",
        )}
        ${this.renderInfoRow(
          "Team",
          "team-label",
          "component-team",
          this.component!.owners?.team || "No team assigned",
        )}
        ${this.renderMaintainersRow()}
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

  private renderMaintainersRow() {
    const maintainers = this.component!.owners?.maintainers;
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
