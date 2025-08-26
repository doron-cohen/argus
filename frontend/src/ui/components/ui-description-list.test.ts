import { expect, fixture, html } from "@open-wc/testing";
import "./ui-description-list.js";

describe("UiDescriptionList", () => {
  it("renders empty when no items", async () => {
    const el = await fixture(html`<ui-description-list></ui-description-list>`);
    await el.updateComplete;

    const container = el.shadowRoot?.querySelector(".u-stack-2");
    expect(container).to.be.null;
  });

  it("renders items correctly", async () => {
    const items = [
      { label: "Name", value: "John Doe" },
      { label: "Email", value: "john@example.com" }
    ];

    const el = await fixture(html`
      <ui-description-list .items=${items}></ui-description-list>
    `);
    await el.updateComplete;

    const labels = el.shadowRoot?.querySelectorAll(".u-font-medium");
    const values = el.shadowRoot?.querySelectorAll(".u-text-primary");

    expect(labels).to.have.length(2);
    expect(values).to.have.length(2);
    expect(labels?.[0]?.textContent?.trim()).to.equal("Name:");
    expect(values?.[0]?.textContent?.trim()).to.equal("John Doe");
  });

  it("renders N/A for empty values", async () => {
    const items = [{ label: "Phone", value: "" }];

    const el = await fixture(html`
      <ui-description-list .items=${items}></ui-description-list>
    `);
    await el.updateComplete;

    const value = el.shadowRoot?.querySelector(".u-text-primary");
    expect(value?.textContent?.trim()).to.equal("N/A");
  });
});
