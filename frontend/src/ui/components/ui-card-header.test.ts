import { expect, fixture, html } from "@open-wc/testing";
import "./ui-card-header.js";
import type { UiCardHeader } from "./ui-card-header";

describe("UiCardHeader", () => {
  it("renders with title and subtitle", async () => {
    const el = await fixture<UiCardHeader>(
      html`<ui-card-header
        title="Test Title"
        subtitle="Test Subtitle"
      ></ui-card-header>`,
    );

    expect(el.title).to.equal("Test Title");
    expect(el.subtitle).to.equal("Test Subtitle");
  });

  it("renders with only title", async () => {
    const el = await fixture<UiCardHeader>(
      html`<ui-card-header title="Test Title"></ui-card-header>`,
    );

    expect(el.title).to.equal("Test Title");
    expect(el.subtitle).to.equal("");
  });

  it("has correct structure", async () => {
    const el = await fixture<UiCardHeader>(
      html`<ui-card-header title="Title" subtitle="Subtitle"></ui-card-header>`,
    );

    const title = el.shadowRoot!.querySelector(".title");
    const subtitle = el.shadowRoot!.querySelector(".subtitle");

    expect(title).to.exist;
    expect(subtitle).to.exist;
  });
});
