# 🛠️ Migration Plan: From Vanilla TypeScript to Lit + Modular SPA Architecture

This guide outlines how to gradually evolve your current setup into a maintainable, testable, and scalable SPA architecture using:

- ✅ Lit for Web Components
- ✅ CSS Custom Properties + Token-based Utility Classes (replaced Tailwind)
- ✅ Nanostores for state
- ✅ Navigo for routing
- ✅ Orval for OpenAPI-based API clients
- ✅ Bun as runtime, bundler, and test runner

---

## ✅ Step 1: Set Up Directory Structure

Establish a clean and scalable layout:

src/
├── api/                  # Typed API clients (OpenAPI-generated)  
├── components/           # Feature-level reusable components  
├── pages/                # Route-aware view components  
├── ui/                   # Design system (buttons, inputs, etc.)  
├── stores/               # Global application-wide state  
├── router/               # Routing config and router-outlet  
├── styles/               # Tailwind and global styles  
├── main.ts               # App entry point  
└── index.html  

---

## ✅ Step 2: Migrate Components to Lit

Convert one component at a time to use Lit’s `LitElement`.

**Before (vanilla):**

class MyComponent extends HTMLElement {
  connectedCallback() {
    this.innerHTML = `<div>Hello</div>`;
  }
}

**After (Lit):**

import { LitElement, html } from 'lit';

class MyComponent extends LitElement {
  render() {
    return html`<div>Hello</div>`;
  }
}
customElements.define('my-component', MyComponent);

- Place this in `components/my-component/index.ts`  
- Import Tailwind globally — class names work in templates  
- Add `data-testid` attributes for testing  

---

## ✅ Step 3: Refactor Tailwind for Global Usage

You already use Tailwind — ensure it's imported in `src/styles/index.css`:

@tailwind base;  
@tailwind components;  
@tailwind utilities;  

Then in `main.ts`:

import './styles/index.css';

Use Tailwind classes inside Lit templates via `class=""`.

---

## ✅ Step 4: Refactor Nanostores per Feature

Create small, focused stores colocated with components or pages.

**Example: `components/post-list/store.ts`**

import { atom } from 'nanostores';

export const $posts = atom({ status: 'idle', data: [] });

export async function fetchPosts() {
  $posts.set({ status: 'loading', data: [] });
  try {
    const res = await fetch('/api/posts');
    const json = await res.json();
    $posts.set({ status: 'success', data: json });
  } catch {
    $posts.set({ status: 'error', data: [] });
  }
}

Use `.subscribe()` or `autorun()` in your components to respond to changes.

---

## ✅ Step 5: Refactor Routing with `<router-outlet>`

Replace string-based tag injection with direct Lit templates.

**`router/outlet.ts`**

import { LitElement, html } from 'lit';  
import Navigo from 'navigo';

export class RouterOutlet extends LitElement {
  router = new Navigo('/');
  currentView = html``;

  connectedCallback() {
    super.connectedCallback();
    this.router
      .on('/', () => this.setView(html`<home-page></home-page>`))
      .on('/about', () => this.setView(html`<about-page></about-page>`))
      .notFound(() => this.setView(html`<not-found-page></not-found-page>`))
      .resolve();
  }

  setView(view) {
    this.currentView = view;
    this.requestUpdate();
  }

  render() {
    return this.currentView;
  }
}
customElements.define('router-outlet', RouterOutlet);

Then use `<router-outlet></router-outlet>` in your HTML entry point.

---

## ✅ Step 6: Move Pages to `src/pages/`

Pages are route-aware components (e.g., `home-page`, `profile-page`).

**Example: `pages/home/index.ts`**

import { LitElement, html } from 'lit';  
import '../../components/post-list';

class HomePage extends LitElement {
  render() {
    return html`
      <h1 class="text-2xl">Home</h1>
      <post-list></post-list>
    `;
  }
}
customElements.define('home-page', HomePage);

Pages may also have a `store.ts` if they manage page-specific data.

---

## ✅ Step 7: Generate Typed API Clients with Orval

1. Install Orval:

bun add -d orval

2. Structure:

src/api/services/users/  
  openapi.yaml  
  orval.config.ts  

3. Example `orval.config.ts`:

export default {
  users: {
    input: './openapi.yaml',
    output: {
      mode: 'tags-split',
      target: './client/',
      client: 'fetch',
    },
  },
};

4. Run:

bunx orval --config src/api/services/users/orval.config.ts

5. Example usage:

import { usersService } from '../../api/services/users/client';

export async function fetchUser(id: string) {
  const res = await usersService.getUserById({ id });
  // update store
}

---

## ✅ Step 8: Create a UI Design System in `/ui`

Reusable, low-level UI components (like buttons, cards) go in `src/ui/`.

**Example: `ui/button/index.ts`**

import { LitElement, html } from 'lit';

class UiButton extends LitElement {
  render() {
    return html`<button class="bg-blue-500 text-white px-4 py-2 rounded">
      <slot></slot>
    </button>`;
  }
}
customElements.define('ui-button', UiButton);

Use Tailwind for styling. Add props and `classMap()` for variants if needed.

---

## ✅ Step 9: Testing

### Unit Tests (Buntest)

- Use `@open-wc/testing`
- Mock stores or APIs
- Use `data-testid` for stable selection

Example:

test('renders greeting', async () => {
  const el = await fixture(html`<my-component></my-component>`);
  expect(el.shadowRoot?.querySelector('[data-testid="greeting"]')?.textContent).toBe('Hello');
});

### E2E Tests (Playwright)

- Place E2E tests in `tests/e2e/*.spec.ts`
- Select with `data-testid`
- Mock APIs if needed using MSW or direct network overrides

---

## ✅ Step 10: Optional Scaffolding with Bun or cargo-generate

- Use Bun scripts to generate components/pages/stores locally
- Use `cargo-generate` + Tera for reusable templates
- Add version tags and `// generated` comments to help with tracking and updates

---

## ✅ Final Summary

You now have:

- 🧱 Modular architecture (`components`, `pages`, `ui`, `stores`, `api`)
- 🧠 Reactive state with `nanostores`
- 🎨 Consistent styling with Tailwind
- 🔌 Typed, documented APIs with Orval
- 🚦 Real routing with Navigo + `<router-outlet>`
- ✅ Unit + E2E testing integration
- 🛠️ Optionally: scaffolding tools to keep it consistent

Take it one step at a time. Build one feature in the new architecture and evolve from there.
