import { expect, fixture, html } from "@open-wc/testing";
import "./ui-info-row.js";
import type { UiInfoRow } from "./ui-info-row";

describe("UiInfoRow", () => {
  it("renders with label and value", async () => {
    const el = await fixture<UiInfoRow>(
      html`<ui-info-row label="Test Label" value="Test Value"></ui-info-row>`
    );

    expect(el.label).to.equal("Test Label");
    expect(el.value).to.equal("Test Value");
  });

  it("has correct structure", async () => {
    const el = await fixture<UiInfoRow>(
      html`<ui-info-row label="Label" value="Value"></ui-info-row>`
    );

    const infoRow = el.shadowRoot!.querySelector(".info-row");
    const label = el.shadowRoot!.querySelector(".label");
    const value = el.shadowRoot!.querySelector(".value");

    expect(infoRow).to.exist;
    expect(label).to.exist;
    expect(value).to.exist;
  });
});
