import { expect, fixture, html } from "@open-wc/testing";
import "./ui-empty-state.js";

describe("UiEmptyState", () => {
  it("renders with default title", async () => {
    const el = await fixture(html`<ui-empty-state></ui-empty-state>`);
    await el.updateComplete;

    const title = el.shadowRoot?.querySelector(".u-text-muted.text-lg");
    expect(title?.textContent?.trim()).to.equal("No data available");
  });

  it("renders with custom title and description", async () => {
    const el = await fixture(html`
      <ui-empty-state
        title="No items found"
        description="Try adjusting your search criteria"
      ></ui-empty-state>
    `);
    await el.updateComplete;

    const title = el.shadowRoot?.querySelector(".u-text-muted.text-lg");
    const description = el.shadowRoot?.querySelector(".u-text-muted.text-sm");

    expect(title?.textContent?.trim()).to.equal("No items found");
    expect(description?.textContent?.trim()).to.equal("Try adjusting your search criteria");
  });

  it("renders without description when not provided", async () => {
    const el = await fixture(html`<ui-empty-state title="Empty"></ui-empty-state>`);
    await el.updateComplete;

    const description = el.shadowRoot?.querySelector(".u-text-muted.text-sm");
    expect(description).to.be.null;
  });
});
