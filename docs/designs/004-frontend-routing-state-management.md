# Design: Frontend Routing and State Management

## Context

As the Argus frontend evolves from simple DOM manipulation to a more structured Web Components architecture, we need lightweight routing and state management solutions. The user wants utilities that are easy to import and use, don't pollute business logic, and remain minimal in both code size and resource usage.

Current state: Basic TypeScript frontend with DOM manipulation, no routing, and ad-hoc global state variables.

## Goals

- **Routing**: Clean URL-based navigation with type-safe route definitions
- **State Management**: Reactive state system that works seamlessly with Web Components
- **Developer Experience**: Simple import/use pattern, expressive API, minimal boilerplate
- **Performance**: Lightweight in bundle size and runtime overhead
- **Maintainability**: Clear separation of concerns, no business logic pollution

## Constraints

- No external framework dependencies (keeping with bare TypeScript approach)
- Must work with standard Web Components
- Bundle size impact should be minimal
- TypeScript-first with full type safety

## Design

### Routing: Navigo Library

**Why Navigo**:
- **Bundle size**: ~3KB minified
- **Perfect for your stack**: Works great with vanilla TypeScript and Web Components
- **Clean API**: Simple route definitions with parameter extraction
- **Active maintenance**: Well-maintained with good TypeScript support
- **No framework dependencies**: Pure vanilla JS/TS

**Installation**:
```bash
bun add navigo
```

**Usage Pattern**:
```typescript
import Navigo from 'navigo';

const router = new Navigo('/');

// Define routes
router.on('/components', () => showComponentsPage());
router.on('/components/:id', (match) => showComponentDetail(match.data.id));
router.on('/sync', () => showSyncPage());

// Start router
router.resolve();

// Navigate programmatically
router.navigate('/components/123');
```

### State Management: Nanostores

**Why Nanostores**:
- **Bundle size**: ~1.5KB minified
- **Framework agnostic**: Works perfectly with Web Components
- **Reactive**: Automatic updates when state changes
- **TypeScript first**: Excellent type safety
- **Active maintenance**: Part of the Astro ecosystem, very well maintained

**Installation**:
```bash
bun add nanostores
```

**Usage Pattern**:
```typescript
import { atom, computed } from 'nanostores';

// Create reactive state
const components = atom<Component[]>([]);
const loading = atom(false);
const error = atom<string | null>(null);

// Computed state
const componentCount = computed(components, (comps) => comps.length);

// Update state
components.set([...newComponents]);
loading.set(true);
error.set('Something went wrong');
```

### Web Component Integration

**Base Component with State Binding**:
```typescript
import { type ReadableAtom } from 'nanostores';

abstract class BaseComponent extends HTMLElement {
  private subscriptions: (() => void)[] = [];
  
  // Automatic cleanup on disconnect
  disconnectedCallback() {
    this.subscriptions.forEach(unsub => unsub());
    this.subscriptions = [];
  }
  
  // Helper for state binding
  protected bindState<T>(store: ReadableAtom<T>, handler: (value: T) => void) {
    const unsub = store.subscribe(handler);
    this.subscriptions.push(unsub);
  }
}
```

**Usage in Components**:
```typescript
import { BaseComponent } from '../utils/base-component.ts';
import { components } from '../stores/app-store.ts';

class ComponentList extends BaseComponent {
  connectedCallback() {
    this.bindState(components, (components) => {
      this.render(components);
    });
  }
  
  private render(components: Component[]) {
    this.innerHTML = components.map(c => 
      `<component-item data-id="${c.id}"></component-item>`
    ).join('');
  }
}
```

### Store Pattern for Complex State

```typescript
import { atom } from 'nanostores';

interface AppState {
  components: Component[];
  loading: boolean;
  error: string | null;
  currentRoute: string;
}

// Individual stores for better granularity
export const components = atom<Component[]>([]);
export const loading = atom(false);
export const error = atom<string | null>(null);
export const currentRoute = atom('');

// Actions
export function setComponents(newComponents: Component[]) {
  components.set(newComponents);
}

export function setLoading(isLoading: boolean) {
  loading.set(isLoading);
}

export function setError(errorMessage: string | null) {
  error.set(errorMessage);
}
```

### File Structure

```
src/
├── utils/
│   └── base-component.ts  # Base Web Component class
├── stores/
│   └── app-store.ts      # Application state stores
├── components/
│   ├── component-list.ts  # Example component
│   └── component-item.ts  # Example component
└── main.ts               # App initialization
```

### Integration Example

**Main Application Setup**:
```typescript
import Navigo from 'navigo';
import { components, setComponents, setLoading } from './stores/app-store.ts';
import './components/component-list.ts';

const router = new Navigo('/');

// Initialize routes
router.on('/', () => router.navigate('/components'));
router.on('/components', () => {
  document.querySelector('#app').innerHTML = '<component-list></component-list>';
  loadComponents();
});

// API integration
async function loadComponents() {
  setLoading(true);
  try {
    const response = await fetch('/api/catalog/v1/components');
    const data = await response.json();
    setComponents(data);
  } catch (err) {
    console.error('Failed to load components:', err);
  } finally {
    setLoading(false);
  }
}

// Start app
router.resolve();
```

## Tradeoffs

**Library vs. Custom**:
- **Pros**: Well-maintained, battle-tested, active community, faster development
- **Cons**: Slightly larger bundle size (~4.5KB total vs. ~5KB custom), less control over API
- **Decision**: Libraries provide better long-term maintainability and community support

**Bundle Size Impact**:
- **Navigo**: ~3KB minified
- **Nanostores**: ~1.5KB minified  
- **Total overhead**: ~4.5KB (very reasonable for the functionality)
- **Alternative**: Custom implementation would be ~5KB but require ongoing maintenance

**API Design**:
- **Navigo**: Simple `.on()` API, good TypeScript support, parameter extraction
- **Nanostores**: Reactive atoms, computed values, excellent TypeScript integration
- **Both**: Clean, expressive APIs that don't pollute business logic

**Maintenance Burden**:
- **Libraries**: Community-maintained, regular updates, bug fixes
- **Custom**: Full control but requires ongoing development effort
- **Decision**: Focus on business logic rather than infrastructure maintenance 