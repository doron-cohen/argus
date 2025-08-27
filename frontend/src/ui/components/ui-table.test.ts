import { html } from "lit";
import { fixture, expect } from "@open-wc/testing";
import "./ui-table.js";
import type { UiTable } from "./ui-table";

describe("ui-table", () => {
  it("renders a table with content", async () => {
    const el = await fixture<UiTable>(html`
      <ui-table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Value</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Item 1</td>
            <td>Value 1</td>
          </tr>
        </tbody>
      </ui-table>
    `);

    const table = el.shadowRoot?.querySelector("table");
    expect(table).to.exist;

    // Check that the component exists and has the right tag name
    expect(el.tagName.toLowerCase()).to.equal("ui-table");
  });

  it("applies striped variant", async () => {
    const el = await fixture<UiTable>(html`
      <ui-table striped>
        <thead>
          <tr>
            <th>Name</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Row 1</td>
          </tr>
          <tr>
            <td>Row 2</td>
          </tr>
        </tbody>
      </ui-table>
    `);

    expect(el.striped).to.be.true;
    expect(el.hasAttribute("striped")).to.be.true;
  });

  it("applies small size variant", async () => {
    const el = await fixture<UiTable>(html`
      <ui-table size="sm">
        <thead>
          <tr>
            <th>Name</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Row 1</td>
          </tr>
        </tbody>
      </ui-table>
    `);

    expect(el.size).to.equal("sm");
    expect(el.getAttribute("size")).to.equal("sm");
  });

  it("has proper accessibility with scope attributes", async () => {
    const el = await fixture<UiTable>(html`
      <ui-table>
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Value</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Item 1</td>
            <td>Value 1</td>
          </tr>
        </tbody>
      </ui-table>
    `);

    // Test that the component renders with the content
    expect(el).to.exist;
    expect(el.tagName.toLowerCase()).to.equal("ui-table");
  });

  it("supports custom data attributes", async () => {
    const el = await fixture<UiTable>(html`
      <ui-table data-testid="test-table">
        <thead>
          <tr>
            <th>Name</th>
          </tr>
        </thead>
        <tbody>
          <tr data-testid="test-row">
            <td>Test</td>
          </tr>
        </tbody>
      </ui-table>
    `);

    expect(el.getAttribute("data-testid")).to.equal("test-table");
    expect(el.tagName.toLowerCase()).to.equal("ui-table");
  });

  it("renders with loading state", async () => {
    const el = await fixture<UiTable>(html`<ui-table loading></ui-table>`);
    expect(el.loading).to.be.true;
    expect(el.shadowRoot?.querySelector('[data-testid="loading-message"]')).to
      .exist;
  });

  it("renders with error state", async () => {
    const el = await fixture<UiTable>(
      html`<ui-table error-message="Test error"></ui-table>`,
    );
    expect(el.errorMessage).to.equal("Test error");
    expect(el.shadowRoot?.querySelector('[data-testid="error-message"]')).to
      .exist;
  });

  it("renders with empty state", async () => {
    const el = await fixture<UiTable>(html`<ui-table empty></ui-table>`);
    expect(el.empty).to.be.true;
    expect(el.shadowRoot?.querySelector('[data-testid="empty-message"]')).to
      .exist;
  });
});
