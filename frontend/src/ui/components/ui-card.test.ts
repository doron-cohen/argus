import { expect, fixture, html } from "@open-wc/testing";
import "./ui-card.js";
import type { UiCard } from "./ui-card.js";

describe("UiCard", () => {
  it("renders with default attributes", async () => {
    const el = await fixture<UiCard>(html`<ui-card></ui-card>`);
    await el.updateComplete;

    expect(el.variant).to.equal("default");
    expect(el.padding).to.equal("md");
    expect(el).to.have.attribute("variant", "default");
    expect(el).to.have.attribute("padding", "md");
  });

  it("renders with different variants", async () => {
    const el = await fixture<UiCard>(
      html`<ui-card variant="elevated"></ui-card>`
    );
    await el.updateComplete;

    expect(el.variant).to.equal("elevated");
    expect(el).to.have.attribute("variant", "elevated");
  });

  it("renders with different padding", async () => {
    const el = await fixture<UiCard>(html`<ui-card padding="lg"></ui-card>`);
    await el.updateComplete;

    expect(el.padding).to.equal("lg");
    expect(el).to.have.attribute("padding", "lg");
  });

  it("renders slot content", async () => {
    const el = await fixture<UiCard>(
      html`<ui-card><div>Test content</div></ui-card>`
    );
    await el.updateComplete;

    const contentSlot = el.shadowRoot?.querySelector(
      ".content slot"
    ) as HTMLSlotElement;
    expect(contentSlot).to.exist;

    const assignedNodes = contentSlot?.assignedNodes();
    expect(assignedNodes).to.have.length(1);
    expect(assignedNodes?.[0].textContent?.trim()).to.equal("Test content");
  });

  it("renders header slot when provided", async () => {
    const el = await fixture<UiCard>(
      html`<ui-card>
        <div slot="header">Header content</div>
        <div>Body content</div>
      </ui-card>`
    );
    await el.updateComplete;

    const header = el.shadowRoot?.querySelector(".header");
    expect(header).to.exist;

    const headerSlot = header?.querySelector(
      'slot[name="header"]'
    ) as HTMLSlotElement;
    expect(headerSlot).to.exist;

    const assignedNodes = headerSlot?.assignedNodes();
    expect(assignedNodes).to.have.length(1);
    expect(assignedNodes?.[0].textContent?.trim()).to.equal("Header content");
  });

  it("renders footer slot when provided", async () => {
    const el = await fixture<UiCard>(
      html`<ui-card>
        <div>Body content</div>
        <div slot="footer">Footer content</div>
      </ui-card>`
    );
    await el.updateComplete;

    const footer = el.shadowRoot?.querySelector(".footer");
    expect(footer).to.exist;

    const footerSlot = footer?.querySelector(
      'slot[name="footer"]'
    ) as HTMLSlotElement;
    expect(footerSlot).to.exist;

    const assignedNodes = footerSlot?.assignedNodes();
    expect(assignedNodes).to.have.length(1);
    expect(assignedNodes?.[0].textContent?.trim()).to.equal("Footer content");
  });

  it("does not render header when no header slot", async () => {
    const el = await fixture<UiCard>(
      html`<ui-card><div>Body content</div></ui-card>`
    );
    await el.updateComplete;

    const header = el.shadowRoot?.querySelector(".header");
    expect(header).to.exist;
    const headerSlot = header?.querySelector(
      'slot[name="header"]'
    ) as HTMLSlotElement;
    expect(headerSlot?.assignedNodes().length).to.equal(0);
  });

  it("does not render footer when no footer slot", async () => {
    const el = await fixture<UiCard>(
      html`<ui-card><div>Body content</div></ui-card>`
    );
    await el.updateComplete;

    const footer = el.shadowRoot?.querySelector(".footer");
    expect(footer).to.exist;
    const footerSlot = footer?.querySelector(
      'slot[name="footer"]'
    ) as HTMLSlotElement;
    expect(footerSlot?.assignedNodes().length).to.equal(0);
  });

  it("updates attributes when properties change", async () => {
    const el = await fixture<UiCard>(html`<ui-card></ui-card>`);

    el.variant = "outlined";
    await el.updateComplete;
    expect(el).to.have.attribute("variant", "outlined");

    el.padding = "sm";
    await el.updateComplete;
    expect(el).to.have.attribute("padding", "sm");
  });
});
