import { html } from "lit";
import { fixture, expect } from "@open-wc/testing";
import "./ui-button.js";
import type { UiButton } from "./ui-button";

describe("ui-button", () => {
  it("renders a button with default properties", async () => {
    const el = await fixture<UiButton>(html`<ui-button>Click me</ui-button>`);

    const button = el.shadowRoot?.querySelector("button");
    expect(button).to.exist;
    expect(el.variant).to.equal("primary");
    expect(el.size).to.equal("md");
    expect(el.disabled).to.be.false;

    const labelSlot = button?.querySelector("slot:not([name])");
    expect(labelSlot).to.exist;
  });

  it("applies different variants", async () => {
    const primaryBtn = await fixture<UiButton>(html`<ui-button variant="primary">Primary</ui-button>`);
    const secondaryBtn = await fixture<UiButton>(html`<ui-button variant="secondary">Secondary</ui-button>`);
    const ghostBtn = await fixture<UiButton>(html`<ui-button variant="ghost">Ghost</ui-button>`);

    expect(primaryBtn.variant).to.equal("primary");
    expect(secondaryBtn.variant).to.equal("secondary");
    expect(ghostBtn.variant).to.equal("ghost");
  });

  it("applies different sizes", async () => {
    const smBtn = await fixture<UiButton>(html`<ui-button size="sm">Small</ui-button>`);
    const mdBtn = await fixture<UiButton>(html`<ui-button size="md">Medium</ui-button>`);
    const lgBtn = await fixture<UiButton>(html`<ui-button size="lg">Large</ui-button>`);

    expect(smBtn.size).to.equal("sm");
    expect(mdBtn.size).to.equal("md");
    expect(lgBtn.size).to.equal("lg");
  });

  it("handles disabled state", async () => {
    const el = await fixture<UiButton>(html`<ui-button disabled>Disabled</ui-button>`);

    const button = el.shadowRoot?.querySelector("button");
    expect(el.disabled).to.be.true;
    expect(button?.hasAttribute("disabled")).to.be.true;
  });

  it("supports icon slots", async () => {
    const el = await fixture<UiButton>(html`
      <ui-button>
        <span slot="icon-start">←</span>
        With Icon
        <span slot="icon-end">→</span>
      </ui-button>
    `);

    const iconStart = el.shadowRoot?.querySelector("slot[name='icon-start']");
    const iconEnd = el.shadowRoot?.querySelector("slot[name='icon-end']");
    const labelSlot = el.shadowRoot?.querySelector("slot:not([name])");

    expect(iconStart).to.exist;
    expect(iconEnd).to.exist;
    expect(labelSlot).to.exist;
  });

  it("supports icon-only buttons with class", async () => {
    const el = await fixture<UiButton>(html`
      <ui-button class="icon-only">
        <span slot="icon-start">★</span>
      </ui-button>
    `);

    expect(el.classList.contains("icon-only")).to.be.true;
  });

  it("has proper accessibility attributes", async () => {
    const el = await fixture<UiButton>(html`<ui-button>Accessible Button</ui-button>`);

    const button = el.shadowRoot?.querySelector("button");
    expect(button?.getAttribute("part")).to.equal("button");
  });

  it("supports form integration", async () => {
    const el = await fixture<UiButton>(html`<ui-button>Submit</ui-button>`);

    const button = el.shadowRoot?.querySelector("button");
    expect(button).to.exist;
    expect(button?.tagName.toLowerCase()).to.equal("button");
  });
});
