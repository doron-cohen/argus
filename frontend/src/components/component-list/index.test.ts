import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import "./index";

function attachHeader() {
  const headerWrapper = document.createElement("div");
  headerWrapper.innerHTML = `
    <div class="px-4 py-5 sm:px-6">
      <h3 class="text-lg leading-6 font-medium text-gray-900" data-testid="components-header">
        Components
      </h3>
    </div>
  `;
  document.body.appendChild(headerWrapper);
  return headerWrapper;
}

async function waitFor(
  predicate: () => boolean,
  timeoutMs = 100,
): Promise<void> {
  const start = Date.now();
  while (!predicate()) {
    if (Date.now() - start > timeoutMs) break;
    await new Promise((r) => setTimeout(r, 0));
  }
}

describe("component-list (Lit)", () => {
  let headerContainer: HTMLElement;
  let originalFetch: any;
  let originalWindowFetch: any;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
    originalWindowFetch = (globalThis as any).window?.fetch;
    headerContainer = attachHeader();
  });

  afterEach(() => {
    document.body.innerHTML = "";
    globalThis.fetch = originalFetch;
    if ((globalThis as any).window)
      (globalThis as any).window.fetch = originalWindowFetch;
  });

  test("shows loading state initially", async () => {
    // Keep fetch pending so component stays in loading state
    const pending = ((..._args: any[]) => new Promise(() => {})) as any;
    globalThis.fetch = pending;
    if ((globalThis as any).window) (globalThis as any).window.fetch = pending;

    const element = document.createElement("component-list");
    document.body.appendChild(element);

    // Allow connectedCallback to run
    await new Promise((r) => setTimeout(r, 0));

    const loading = element.querySelector('[data-testid="loading-message"]');
    expect(loading).toBeTruthy();
    const header = document.querySelector('[data-testid="components-header"]');
    expect(header?.textContent?.trim()).toBe("Components");
  });

  test("renders empty state and updates header count", async () => {
    const okEmpty = (() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve([]),
      }) as any) as any;
    globalThis.fetch = okEmpty;
    if ((globalThis as any).window) (globalThis as any).window.fetch = okEmpty;

    const element = document.createElement("component-list");
    document.body.appendChild(element);

    // Wait for fetch + render
    await new Promise((r) => setTimeout(r, 0));
    await new Promise((r) => setTimeout(r, 0));

    const empty = element.querySelector(
      '[data-testid="no-components-message"]',
    );
    expect(empty).toBeTruthy();

    const header = document.querySelector('[data-testid="components-header"]');
    expect(header?.textContent?.trim()).toBe("Components (0)");
  });

  test("renders error state", async () => {
    const errResp = (() =>
      Promise.resolve({
        ok: false,
        status: 500,
        statusText: "Server Error",
        json: () => Promise.resolve({ error: "Boom" }),
      }) as any) as any;
    globalThis.fetch = errResp;
    if ((globalThis as any).window) (globalThis as any).window.fetch = errResp;

    const element = document.createElement("component-list");
    document.body.appendChild(element);

    await new Promise((r) => setTimeout(r, 0));
    await new Promise((r) => setTimeout(r, 0));

    const errorEl = element.querySelector('[data-testid="error-message"]');
    expect(errorEl).toBeTruthy();
    expect(errorEl?.textContent || "").toContain("Error:");

    const header = document.querySelector('[data-testid="components-header"]');
    expect(header?.textContent?.trim()).toBe("Components");
  });

  test("renders rows and updates header count", async () => {
    const data = [
      {
        id: "a",
        name: "A",
        description: "desc",
        owners: { maintainers: ["m"], team: "t" },
      },
      {
        id: "b",
        name: "B",
        description: "desc",
        owners: { maintainers: [], team: "t" },
      },
    ];
    const element = document.createElement("component-list") as any;
    // Prevent auto-fetch in connectedCallback for this test
    (element as any).loadComponents = async () => {};
    document.body.appendChild(element);

    // Set state directly to avoid environment-specific fetch behavior
    element.components = data;
    (element as any).isLoading = false;
    (element as any).error = null;
    element.requestUpdate?.();
    await new Promise((r) => setTimeout(r, 0));
    if (element.updateComplete) await element.updateComplete;
    element.updateHeader?.();

    await waitFor(
      () =>
        (element as HTMLElement).querySelectorAll(
          '[data-testid="component-row"]',
        ).length === 2,
      500,
    );
    const rows = (element as HTMLElement).querySelectorAll(
      '[data-testid="component-row"]',
    );
    expect(rows.length).toBe(2);

    const header = document.querySelector('[data-testid="components-header"]');
    expect(header?.textContent?.trim()).toBe("Components (2)");
  });
});
