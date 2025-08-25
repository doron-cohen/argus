import { expect, fixture, html } from "@open-wc/testing";
import "./page-container.js";
import type { UiPageContainer } from "./page-container";

describe("UiPageContainer", () => {
  it("renders with default attributes", async () => {
    const el = await fixture<UiPageContainer>(
      html`<ui-page-container></ui-page-container>`,
    );
    await el.updateComplete;

    expect(el.padding).to.equal("md");
    expect(el.maxWidth).to.equal("lg");
    expect(el).to.have.attribute("padding", "md");
    expect(el).to.have.attribute("max-width", "lg");
  });

  it("renders with custom padding", async () => {
    const el = await fixture<UiPageContainer>(
      html`<ui-page-container padding="lg"></ui-page-container>`,
    );
    await el.updateComplete;

    expect(el.padding).to.equal("lg");
    expect(el).to.have.attribute("padding", "lg");
  });

  it("renders with custom max width", async () => {
    const el = await fixture<UiPageContainer>(
      html`<ui-page-container max-width="xl"></ui-page-container>`,
    );
    await el.updateComplete;

    expect(el.maxWidth).to.equal("xl");
    expect(el).to.have.attribute("max-width", "xl");
  });

  it("renders slot content", async () => {
    const el = await fixture<UiPageContainer>(
      html`<ui-page-container><div>Test content</div></ui-page-container>`,
    );
    await el.updateComplete;

    const slot = el.shadowRoot?.querySelector("slot") as HTMLSlotElement;
    expect(slot).to.exist;

    const assignedNodes = slot?.assignedNodes();
    expect(assignedNodes).to.have.length(1);
    expect(assignedNodes?.[0].textContent?.trim()).to.equal("Test content");
  });

  it("updates attributes when properties change", async () => {
    const el = await fixture<UiPageContainer>(
      html`<ui-page-container></ui-page-container>`,
    );

    el.padding = "sm";
    await el.updateComplete;
    expect(el).to.have.attribute("padding", "sm");

    el.maxWidth = "full";
    await el.updateComplete;
    expect(el).to.have.attribute("max-width", "full");
  });
});
