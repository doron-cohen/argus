import { expect, fixture, html } from "@open-wc/testing";
import "./ui-page-header.js";
import type { UiPageHeader } from "./ui-page-header.js";

describe("UiPageHeader", () => {
  it("renders with default attributes", async () => {
    const el = await fixture<UiPageHeader>(
      html`<ui-page-header></ui-page-header>`
    );
    await el.updateComplete;
    await el.updateComplete;
    expect(el.title).to.equal("");
    expect(el.description).to.equal("");
    expect(el.size).to.equal("md");
  });

  it("renders title and description", async () => {
    const el = await fixture<UiPageHeader>(
      html`<ui-page-header
        title="Test Title"
        description="Test Description"
      ></ui-page-header>`
    );

    await el.updateComplete;
    expect(el.title).to.equal("Test Title");
    expect(el.description).to.equal("Test Description");

    const title = el.shadowRoot?.querySelector(".title");
    const description = el.shadowRoot?.querySelector(".description");

    expect(title?.textContent?.trim()).to.equal("Test Title");
    expect(description?.textContent?.trim()).to.equal("Test Description");
  });

  it("renders with different sizes", async () => {
    const el = await fixture<UiPageHeader>(
      html`<ui-page-header title="Test" size="lg"></ui-page-header>`
    );
    await el.updateComplete;

    expect(el.size).to.equal("lg");
    expect(el).to.have.attribute("size", "lg");
  });

  it("renders actions slot", async () => {
    const el = await fixture<UiPageHeader>(
      html`<ui-page-header title="Test">
        <button slot="actions">Action</button>
      </ui-page-header>`
    );
    await el.updateComplete;

    const actionsSlot = el.shadowRoot?.querySelector(
      'slot[name="actions"]'
    ) as HTMLSlotElement;
    expect(actionsSlot).to.exist;

    const assignedNodes = actionsSlot?.assignedNodes();
    expect(assignedNodes).to.have.length(1);
    expect(assignedNodes?.[0].textContent?.trim()).to.equal("Action");
  });

  it("does not render title when empty", async () => {
    const el = await fixture<UiPageHeader>(
      html`<ui-page-header description="Test Description"></ui-page-header>`
    );
    await el.updateComplete;

    const title = el.shadowRoot?.querySelector(".title");
    expect(title).to.not.exist;
  });

  it("does not render description when empty", async () => {
    const el = await fixture<UiPageHeader>(
      html`<ui-page-header title="Test Title"></ui-page-header>`
    );
    await el.updateComplete;

    const description = el.shadowRoot?.querySelector(".description");
    expect(description).to.not.exist;
  });

  it("updates content when properties change", async () => {
    const el = await fixture<UiPageHeader>(
      html`<ui-page-header
        title="Old Title"
        description="Old Description"
      ></ui-page-header>`
    );

    await el.updateComplete;
    el.title = "New Title";
    el.description = "New Description";
    await el.updateComplete;

    const title = el.shadowRoot?.querySelector(".title");
    const description = el.shadowRoot?.querySelector(".description");

    expect(title?.textContent?.trim()).to.equal("New Title");
    expect(description?.textContent?.trim()).to.equal("New Description");
  });
});
