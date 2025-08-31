import { expect, fixture, html } from "@open-wc/testing";
import "./metadata";

// Import UI components used by component-metadata
import "../../ui/components/ui-card.js";
import "../../ui/components/ui-card-header.js";
import "../../ui/components/ui-info-row.js";

import type { Component } from "../../api/services/components/client";
import type { ComponentMetadata } from "./metadata";

describe("component-metadata", () => {
  const mockComponent: Component = {
    id: "test-component",
    name: "Test Component",
    description: "This is a test component",
    owners: {
      maintainers: ["john.doe", "jane.smith"],
      team: "Platform Team",
    },
  };

  it("component is defined", () => {
    expect(customElements.get("component-metadata")).to.exist;
  });

  it("renders empty state when no component is provided", async () => {
    const el = await fixture(html` <component-metadata></component-metadata> `);

    expect(el.shadowRoot?.textContent).to.include(
      "No component data available",
    );
  });

  it("renders component metadata when component is provided", async () => {
    const el = await fixture<ComponentMetadata>(html`
      <component-metadata .component=${mockComponent}></component-metadata>
    `);
    await el.updateComplete;

    const name = el.shadowRoot?.querySelector(
      '[title-data-testid="component-name"]',
    );
    expect(name).to.exist;
    expect(name?.getAttribute("title")).to.equal("Test Component");

    const description = el.shadowRoot?.querySelector(
      '[value-data-testid="component-description"]',
    );
    expect(description).to.exist;
    expect(description?.getAttribute("value")).to.equal(
      "This is a test component",
    );
  });

  it("renders component id in subtitle", async () => {
    const el = await fixture<ComponentMetadata>(html`
      <component-metadata .component=${mockComponent}></component-metadata>
    `);
    await el.updateComplete;

    const id = el.shadowRoot?.querySelector(
      '[subtitle-data-testid="component-id"]',
    );
    expect(id).to.exist;
    expect(id?.getAttribute("subtitle")).to.equal("ID: test-component");
  });

  it("renders team information", async () => {
    const el = await fixture<ComponentMetadata>(html`
      <component-metadata .component=${mockComponent}></component-metadata>
    `);
    await el.updateComplete;

    const team = el.shadowRoot?.querySelector(
      '[value-data-testid="component-team"]',
    );
    expect(team).to.exist;
    expect(team?.getAttribute("value")).to.equal("Platform Team");
  });

  it("renders maintainers as comma-separated list", async () => {
    const el = await fixture<ComponentMetadata>(html`
      <component-metadata .component=${mockComponent}></component-metadata>
    `);
    await el.updateComplete;

    const maintainers = el.shadowRoot?.querySelector(
      '[value-data-testid="component-maintainers"]',
    );
    expect(maintainers).to.exist;
    expect(maintainers?.getAttribute("value")).to.equal("john.doe, jane.smith");
  });

  it("renders no maintainers assigned when empty", async () => {
    const componentWithoutMaintainers: Component = {
      ...mockComponent,
      owners: { maintainers: [], team: "Platform Team" },
    };

    const el = await fixture<ComponentMetadata>(html`
      <component-metadata
        .component=${componentWithoutMaintainers}
      ></component-metadata>
    `);
    await el.updateComplete;

    const maintainers = el.shadowRoot?.querySelector(
      '[value-data-testid="component-maintainers"]',
    );
    expect(maintainers).to.exist;
    expect(maintainers?.getAttribute("value")).to.equal(
      "No maintainers assigned",
    );
  });

  it("renders no team assigned when team is missing", async () => {
    const componentWithoutTeam: Component = {
      ...mockComponent,
      owners: { maintainers: ["john.doe"] },
    };

    const el = await fixture<ComponentMetadata>(html`
      <component-metadata
        .component=${componentWithoutTeam}
      ></component-metadata>
    `);
    await el.updateComplete;

    const team = el.shadowRoot?.querySelector(
      '[value-data-testid="component-team"]',
    );
    expect(team).to.exist;
    expect(team?.getAttribute("value")).to.equal("No team assigned");
  });

  it("renders no description available when description is missing", async () => {
    const componentWithoutDescription: Component = {
      ...mockComponent,
      description: undefined,
    };

    const el = await fixture<ComponentMetadata>(html`
      <component-metadata
        .component=${componentWithoutDescription}
      ></component-metadata>
    `);
    await el.updateComplete;

    const description = el.shadowRoot?.querySelector(
      '[value-data-testid="component-description"]',
    );
    expect(description).to.exist;
    expect(description?.getAttribute("value")).to.equal(
      "No description available",
    );
  });
});
