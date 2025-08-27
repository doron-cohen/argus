import { expect, fixture, html } from "@open-wc/testing";
import "./index";

describe("component-list", () => {
  const mockComponents = [
    {
      id: "a",
      name: "Component A",
      description: "Description A",
      owners: { maintainers: ["user1"], team: "Team A" },
    },
    {
      id: "b",
      name: "Component B",
      description: "Description B",
      owners: { maintainers: ["user2"], team: "Team B" },
    },
  ];

  it("component is defined", () => {
    expect(customElements.get("component-list")).to.exist;
  });

  it("shows loading state when isLoading is true", async () => {
    const el = await fixture(html`
      <component-list
        .isLoading=${true}
        .components=${[]}
        .error=${null}
      ></component-list>
    `);

    const uiTable = el.shadowRoot?.querySelector("ui-table");
    const loading = uiTable?.shadowRoot?.querySelector(
      '[data-testid="loading-message"]',
    );
    expect(loading).to.exist;
    expect(loading?.textContent?.trim()).to.equal("Loading components...");
  });

  it("renders empty state when no components", async () => {
    const el = await fixture(html`
      <component-list
        .components=${[]}
        .isLoading=${false}
        .error=${null}
      ></component-list>
    `);

    const uiTable = el.shadowRoot?.querySelector("ui-table");
    const empty = uiTable?.shadowRoot?.querySelector(
      '[data-testid="empty-message"]',
    );
    expect(empty).to.exist;
    expect(empty?.textContent?.trim()).to.equal("No components found");
  });

  it("renders error state when error is provided", async () => {
    const el = await fixture(html`
      <component-list
        .components=${[]}
        .isLoading=${false}
        .error=${"HTTP 500"}
      ></component-list>
    `);

    const uiTable = el.shadowRoot?.querySelector("ui-table");
    const errorEl = uiTable?.shadowRoot?.querySelector(
      '[data-testid="error-message"]',
    );
    expect(errorEl).to.exist;
    expect(errorEl?.textContent?.trim()).to.equal("HTTP 500");
  });

  it("renders component rows when components are provided", async () => {
    const el = await fixture(html`
      <component-list
        .components=${mockComponents}
        .isLoading=${false}
        .error=${null}
      ></component-list>
    `);

    const rows = el.shadowRoot?.querySelectorAll(
      '[data-testid="component-row"]',
    );
    expect(rows).to.have.length(2);
  });

  it("renders component names correctly", async () => {
    const el = await fixture(html`
      <component-list
        .components=${mockComponents}
        .isLoading=${false}
        .error=${null}
      ></component-list>
    `);

    const rows = el.shadowRoot?.querySelectorAll(
      '[data-testid="component-row"]',
    );
    expect(rows).to.have.length(2);

    const firstRow = rows?.[0];
    const name = firstRow?.querySelector('[data-testid="component-name"]');
    expect(name?.textContent?.trim()).to.equal("Component A");
  });

  it("renders component descriptions correctly", async () => {
    const el = await fixture(html`
      <component-list
        .components=${mockComponents}
        .isLoading=${false}
        .error=${null}
      ></component-list>
    `);

    const rows = el.shadowRoot?.querySelectorAll(
      '[data-testid="component-row"]',
    );
    const firstRow = rows?.[0];
    const description = firstRow?.querySelector(
      '[data-testid="component-description"]',
    );
    expect(description?.textContent?.trim()).to.equal("Description A");
  });

  it("renders team information correctly", async () => {
    const el = await fixture(html`
      <component-list
        .components=${mockComponents}
        .isLoading=${false}
        .error=${null}
      ></component-list>
    `);

    const rows = el.shadowRoot?.querySelectorAll(
      '[data-testid="component-row"]',
    );
    const firstRow = rows?.[0];
    const team = firstRow?.querySelector('[data-testid="component-team"]');
    expect(team?.textContent?.trim()).to.equal("Team A");
  });

  it("renders maintainers correctly", async () => {
    const el = await fixture(html`
      <component-list
        .components=${mockComponents}
        .isLoading=${false}
        .error=${null}
      ></component-list>
    `);

    const rows = el.shadowRoot?.querySelectorAll(
      '[data-testid="component-row"]',
    );
    const firstRow = rows?.[0];
    const maintainers = firstRow?.querySelector(
      '[data-testid="component-maintainers"]',
    );
    expect(maintainers?.textContent?.trim()).to.equal("user1");
  });

  it("handles components without maintainers", async () => {
    const componentsWithoutMaintainers = [
      {
        id: "c",
        name: "Component C",
        description: "Description C",
        owners: { team: "Team C" },
      },
    ];

    const el = await fixture(html`
      <component-list
        .components=${componentsWithoutMaintainers}
        .isLoading=${false}
        .error=${null}
      ></component-list>
    `);

    const rows = el.shadowRoot?.querySelectorAll(
      '[data-testid="component-row"]',
    );
    const row = rows?.[0];
    const maintainers = row?.querySelector(
      '[data-testid="component-maintainers"]',
    );
    expect(maintainers?.textContent?.trim()).to.equal("");
  });
});
