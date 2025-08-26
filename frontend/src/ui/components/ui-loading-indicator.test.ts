import { expect, fixture, html } from "@open-wc/testing";
import "./ui-loading-indicator.js";

describe("UiLoadingIndicator", () => {
  it("renders with default message", async () => {
    const el = await fixture(html`<ui-loading-indicator></ui-loading-indicator>`);
    await el.updateComplete;

    const message = el.shadowRoot?.querySelector(".u-text-muted");
    expect(message?.textContent?.trim()).to.equal("Loading...");
  });

  it("renders with custom message", async () => {
    const el = await fixture(html`<ui-loading-indicator message="Please wait..."></ui-loading-indicator>`);
    await el.updateComplete;

    const message = el.shadowRoot?.querySelector(".u-text-muted");
    expect(message?.textContent?.trim()).to.equal("Please wait...");
  });

  it("renders with different sizes", async () => {
    const el = await fixture(html`<ui-loading-indicator size="lg"></ui-loading-indicator>`);
    await el.updateComplete;

    const spinner = el.shadowRoot?.querySelector(".animate-spin");
    expect(spinner?.classList.contains("h-8")).to.be.true;
    expect(spinner?.classList.contains("w-8")).to.be.true;
  });
});
