import { expect, fixture, html } from "@open-wc/testing";
import "./ui-alert.js";
import type { UiAlert } from "./ui-alert";

describe("UiAlert", () => {
  it("renders with default properties", async () => {
    const el = await fixture<UiAlert>(html`<ui-alert></ui-alert>`);

    expect(el).to.have.attribute("variant", "info");
  });

  it("renders with title and message", async () => {
    const el = await fixture<UiAlert>(
      html`<ui-alert title="Test Title" message="Test message"></ui-alert>`,
    );

    expect(el.title).to.equal("Test Title");
    expect(el.message).to.equal("Test message");
  });

  it("renders with error variant", async () => {
    const el = await fixture<UiAlert>(
      html`<ui-alert variant="error"></ui-alert>`,
    );

    expect(el).to.have.attribute("variant", "error");
  });

  it("renders dismissible alert", async () => {
    const el = await fixture<UiAlert>(html`<ui-alert dismissible></ui-alert>`);

    const dismissButton = el.shadowRoot!.querySelector(".dismiss-button");
    expect(dismissButton).to.exist;
  });

  it("has correct structure", async () => {
    const el = await fixture<UiAlert>(
      html`<ui-alert title="Title" message="Message"></ui-alert>`,
    );

    const alertContent = el.shadowRoot!.querySelector(".alert-content");
    const alertTitle = el.shadowRoot!.querySelector(".alert-title");
    const alertMessage = el.shadowRoot!.querySelector(".alert-message");

    expect(alertContent).to.exist;
    expect(alertTitle).to.exist;
    expect(alertMessage).to.exist;
  });
});
