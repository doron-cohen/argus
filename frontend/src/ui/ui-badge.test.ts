import { expect, fixture, html } from "@open-wc/testing";
import "./ui-badge";
import type { UiBadge } from "./ui-badge";

describe("ui-badge", () => {
  function getClasses(el: UiBadge) {
    return (el.className || "").split(/\s+/);
  }

  function getTextContent(el: UiBadge) {
    return el.shadowRoot?.textContent?.trim() || "";
  }

  it("renders pass variant with green classes and check icon", async () => {
    const el = await fixture<UiBadge>(
      html`<ui-badge status="pass"></ui-badge>`,
    );
    expect(getClasses(el)).to.include("pass");
    expect(getTextContent(el)).to.include("pass");
  });

  it("renders fail/error/unknown with status class", async () => {
    for (const status of ["fail", "error", "unknown"]) {
      const el = await fixture<UiBadge>(
        html`<ui-badge status=${status}></ui-badge>`,
      );
      expect(getClasses(el)).to.include(status);
      expect(getTextContent(el)).to.include(status);
    }
  });

  it("renders disabled/skipped with status class", async () => {
    for (const status of ["disabled", "skipped"]) {
      const el = await fixture<UiBadge>(
        html`<ui-badge status=${status}></ui-badge>`,
      );
      expect(getClasses(el)).to.include(status);
    }
  });

  it("renders completed with status class", async () => {
    const el = await fixture<UiBadge>(
      html`<ui-badge status="completed"></ui-badge>`,
    );
    expect(getClasses(el)).to.include("completed");
  });

  it("renders default for unknown value", async () => {
    const el = await fixture<UiBadge>(
      html`<ui-badge status="weird"></ui-badge>`,
    );
    expect(getClasses(el)).to.include("default");
    expect(getTextContent(el)).to.include("weird");
  });

  it("escapes label text", async () => {
    const el = await fixture<UiBadge>(
      html`<ui-badge .status=${'<img src=x onerror="1">'}></ui-badge>`,
    );
    expect(el.shadowRoot?.innerHTML).to.not.include("<img");
    expect(getTextContent(el)).to.include(
      "&lt;img src=x onerror=&quot;1&quot;&gt;",
    );
  });

  it("is case-insensitive for classes but preserves label", async () => {
    const el = await fixture<UiBadge>(
      html`<ui-badge status="PASS"></ui-badge>`,
    );
    expect(getClasses(el)).to.include("pass");
    expect(getTextContent(el)).to.equal("PASS");
  });

  it("renders default when no status attribute provided", async () => {
    const el = await fixture<UiBadge>(html`<ui-badge></ui-badge>`);
    expect(getClasses(el)).to.include("default");
    expect(getTextContent(el)).to.equal("default");
  });
});
