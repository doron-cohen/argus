import { expect, fixture, html } from "@open-wc/testing";
import "./ui-spinner.js";
import type { UiSpinner } from "./ui-spinner";

describe("UiSpinner", () => {
  it("renders with default properties", async () => {
    const el = await fixture<UiSpinner>(html`<ui-spinner></ui-spinner>`);

    expect(el).to.have.attribute("size", "md");
    expect(el).to.have.attribute("color", "primary");
  });

  it("renders with custom size", async () => {
    const el = await fixture<UiSpinner>(
      html`<ui-spinner size="lg"></ui-spinner>`,
    );

    expect(el).to.have.attribute("size", "lg");
  });

  it("renders with custom color", async () => {
    const el = await fixture<UiSpinner>(
      html`<ui-spinner color="secondary"></ui-spinner>`,
    );

    expect(el).to.have.attribute("color", "secondary");
  });

  it("has correct structure", async () => {
    const el = await fixture<UiSpinner>(html`<ui-spinner></ui-spinner>`);

    const spinner = el.shadowRoot!.querySelector(".spinner");
    expect(spinner).to.exist;
  });
});
