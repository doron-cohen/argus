import { expect, fixture, html } from "@open-wc/testing";
import "./ui-icon.js";
import type { UiIcon } from "./ui-icon";

describe("UiIcon", () => {
  it("renders with default attributes", async () => {
    const el = await fixture<UiIcon>(html`<ui-icon></ui-icon>`);
    await el.updateComplete;

    expect(el.name).to.equal("check");
    expect(el.size).to.equal("sm");
    expect(el).to.have.attribute("name", "check");
    expect(el).to.have.attribute("size", "sm");
  });

  it("renders with custom name", async () => {
    const el = await fixture<UiIcon>(html`<ui-icon name="x"></ui-icon>`);
    await el.updateComplete;

    expect(el.name).to.equal("x");
    expect(el).to.have.attribute("name", "x");
  });

  it("renders with custom size", async () => {
    const el = await fixture<UiIcon>(html`<ui-icon size="lg"></ui-icon>`);
    await el.updateComplete;

    expect(el.size).to.equal("lg");
    expect(el).to.have.attribute("size", "lg");
  });

  it("renders SVG content", async () => {
    const el = await fixture<UiIcon>(html`<ui-icon name="check"></ui-icon>`);
    await el.updateComplete;

    const svg = el.shadowRoot?.querySelector("svg");
    expect(svg).to.exist;
    expect(svg?.getAttribute("viewBox")).to.equal("0 0 20 20");

    const path = svg?.querySelector("path");
    expect(path).to.exist;
    expect(path?.getAttribute("fill-rule")).to.equal("evenodd");
  });

  it("renders different icons", async () => {
    const icons = ["check", "x", "warning", "circle-check"] as const;

    for (const iconName of icons) {
      const el = await fixture<UiIcon>(
        html`<ui-icon name=${iconName}></ui-icon>`,
      );
      await el.updateComplete;

      expect(el.name).to.equal(iconName);
      expect(el).to.have.attribute("name", iconName);

      const svg = el.shadowRoot?.querySelector("svg");
      expect(svg).to.exist;
    }
  });

  it("handles invalid icon name gracefully", async () => {
    const el = await fixture<UiIcon>(
      html`<ui-icon name="invalid-icon"></ui-icon>`,
    );
    await el.updateComplete;

    const svg = el.shadowRoot?.querySelector("svg");
    expect(svg).to.not.exist;
  });
});
