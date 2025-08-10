// Minimal helpers for testing Lit components with Bun + happy-dom

export async function flushPromises(): Promise<void> {
  await Promise.resolve();
  await Promise.resolve();
  await new Promise((r) => setTimeout(r, 0));
}

export async function waitFor(
  predicate: () => boolean,
  timeoutMs = 200,
  intervalMs = 0,
): Promise<void> {
  const start = Date.now();
  // Prime microtasks first
  await flushPromises();
  while (!predicate()) {
    if (Date.now() - start > timeoutMs) break;
    // Yield to event loop
    await new Promise((r) => setTimeout(r, intervalMs));
  }
}

export function renderElement<T extends HTMLElement>(
  tagName: string,
  attributes?: Record<string, unknown>,
  container: HTMLElement = document.body,
): T {
  const el = document.createElement(tagName) as T;
  if (attributes) {
    for (const [key, value] of Object.entries(attributes)) {
      if (value === true) {
        el.setAttribute(key, "");
      } else if (value !== false && value != null) {
        el.setAttribute(key, String(value));
      }
    }
  }
  container.appendChild(el);
  return el;
}

export function attachHeader(): HTMLElement {
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
