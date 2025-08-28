# Design: Web UI Infrastructure

## Context
**Users**: Developers and maintainers who need to view the component catalog
**Pain**: Currently no web interface to browse components - only API access available
**Solution**: A lightweight web UI that displays components in a simple, functional interface

## Goals
- Provide a basic web interface for viewing the component catalog
- Establish infrastructure for future UI development
- Keep implementation simple with minimal dependencies
- Support both unit and end-to-end testing
- Integrate with existing Makefile workflow

## Constraints
- Must be lightweight and simple (no complex frameworks)
- Should integrate with existing Go backend
- Must support testing infrastructure
- Should follow established URL scheme from design 002
- Minimize npm dependencies and package management complexity

## Design

### Technology Stack
- **Bun**: Fast, all-in-one JavaScript runtime and package manager (replaces npm/yarn)
- **TypeScript**: Type-safe JavaScript with plain web components approach
- **CSS Custom Properties + Token-based Utility Classes**: Semantic design system with theme support (replaced Tailwind)
- **Playwright**: End-to-end testing framework with comprehensive test coverage

### Package Management Strategy
**Bun over npm/yarn**: Based on research, Bun offers significant advantages:
- **Speed**: Up to 30x faster installation than npm
- **All-in-one**: Runtime, bundler, test runner, and package manager
- **Minimal dependencies**: Built-in tools reduce external package needs
- **Go integration**: Similar performance philosophy to our backend

### Directory Structure
```
frontend/
├── src/
│   ├── main.ts
│   ├── utils.ts
│   └── styles.css
├── tests/
│   ├── unit/
│   │   ├── escapeHtml.test.ts
│   │   └── security.test.ts
│   ├── e2e/
│   │   ├── components.spec.ts
│   │   ├── sync.spec.ts
│   │   └── types.ts
│   └── tsconfig.json
├── dist/
│   ├── main.js
│   └── styles.css
├── index.html
├── package.json
├── playwright.config.ts
└── bun.lockb
```

### URL Structure
**Backend serves frontend from root**:
- `/` - Components list page (main page)
- `/api/*` - API endpoints (existing)
- Frontend files served from `/` route by Go backend

### Components List Page
**Simple table view showing:**
- Component name
- Component ID
- Description
- Team
- Maintainers

**TypeScript Implementation:**
```typescript
// main.ts
interface Component {
  id: string;
  name: string;
  description: string;
  owners: {
    maintainers: string[];
    team: string;
  };
}

let components: Component[] = [];
let isLoading = true;
let error: string | null = null;

async function fetchComponents(): Promise<void> {
  try {
    isLoading = true;
    error = null;
    
    const response = await fetch("/api/catalog/v1/components");
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    components = await response.json();
  } catch (err) {
    error = err instanceof Error ? err.message : "Failed to fetch components";
  } finally {
    isLoading = false;
    renderComponents();
  }
}

function renderComponents(): void {
  const tbody = document.getElementById("components-tbody");
  const countSpan = document.getElementById("component-count");
  
  if (!tbody || !countSpan) return;
  
  countSpan.textContent = components.length.toString();
  
  if (isLoading) {
    tbody.innerHTML = `<tr><td colspan="5" class="text-center">Loading...</td></tr>`;
    return;
  }
  
  if (error) {
    tbody.innerHTML = `<tr><td colspan="5" class="text-center text-red-500">Error: ${escapeHtml(error)}</td></tr>`;
    return;
  }
  
  tbody.innerHTML = components.map(comp => `
    <tr class="hover:bg-gray-50">
      <td class="px-6 py-4">${escapeHtml(comp.name)}</td>
      <td class="px-6 py-4">${escapeHtml(comp.id)}</td>
      <td class="px-6 py-4">${escapeHtml(comp.description)}</td>
      <td class="px-6 py-4">${escapeHtml(comp.owners?.team || "")}</td>
      <td class="px-6 py-4">${escapeHtml(comp.owners?.maintainers?.join(", ") || "")}</td>
    </tr>
  `).join("");
}
```

### Styling Approach
**CSS Custom Properties + Token-based Utility Classes:**
- Use semantic CSS custom properties for theming (light/dark modes)
- Token-backed utility classes (`u-*`) for consistent styling
- Lit CSS-in-JS for component encapsulation
- Minimal custom CSS for component-specific needs
- Focus on readability and usability over aesthetics
- Responsive design for different screen sizes

### Testing Strategy

#### Unit Tests (Bun Test)
**Test TypeScript utilities and security functions:**
```typescript
// escapeHtml.test.ts
import { test, expect } from "bun:test";
import { escapeHtml } from "../src/utils.ts";

test("escapeHtml should escape HTML special characters", () => {
  expect(escapeHtml("<script>alert('xss')</script>"))
    .toBe("&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;");
});

test("escapeHtml should handle empty string", () => {
  expect(escapeHtml("")).toBe("");
});
```

#### End-to-End Tests (Playwright)
**Comprehensive test coverage with real application flow:**
```typescript
// components.spec.ts
import { test, expect } from "@playwright/test";

test.describe("Component Catalog - Real Application Flow", () => {
  test("should display all test components from sync process", async ({ page }) => {
    await page.goto("/");
    
    // Wait for components to load and verify count
    await expect(page.getByTestId("component-row")).toHaveCount(4);
    await expect(page.getByTestId("components-header")).toContainText("Components (4)");
    
    // Verify specific components
    const authServiceRow = page.getByTestId("component-row").filter({ hasText: "Authentication Service" });
    await expect(authServiceRow).toHaveCount(1);
    await expect(authServiceRow.getByTestId("component-team")).toContainText("Security Team");
  });
  
  test("should verify real API responses match frontend display", async ({ page }) => {
    // Get components via API
    const apiResponse = await page.request.get("http://localhost:8080/api/catalog/v1/components");
    expect(apiResponse.ok()).toBeTruthy();
    
    const components = await apiResponse.json();
    expect(components).toHaveLength(4);
    
    // Verify frontend displays same data
    await page.goto("/");
    await expect(page.getByTestId("component-row")).toHaveCount(4);
  });
});
```

### Build Pipeline & Makefile Integration
**Makefile targets for frontend workflows:**
```makefile
# Frontend targets
frontend/install:
	cd frontend && bun install

frontend/dev:
	cd frontend && bun run dev

frontend/build:
	cd frontend && bun run build:prod

frontend/test:
	cd frontend && bun run type-check

frontend/test-unit:
	cd frontend && bun run test:unit

frontend/test-e2e:
	cd frontend && bun run test:e2e

frontend/lint:
	cd frontend && bun run type-check

frontend/ci: frontend/install frontend/lint frontend/test frontend/build frontend/validate-build frontend/test-e2e-ci

# Combined targets
all: backend/gen-all backend/go-mod-tidy frontend/build

ci: backend/ci frontend/ci
```

**Bun-based development workflow:**
1. **Install**: `make frontend/install` - Fast dependency installation
2. **Development**: `bun run dev` - Hot reload with built-in dev server
3. **Testing**: `bun run test:unit` - Built-in test runner for unit tests
4. **E2E Testing**: `bun run test:e2e` - Playwright tests
5. **Build**: `bun run build:prod` - Production build with built-in bundler
6. **Type Check**: `bun run type-check` - TypeScript validation

### Integration with Backend
**Go backend serves frontend files:**
- Frontend build output served from `/` route in Go
- API routes remain at `/api/*` (design 002 compliance)
- Development: Bun dev server with API proxy to Go backend
- Production: Go serves static files from `frontend/dist/`

## Tradeoffs

### Framework Choice
- **Chosen**: Plain TypeScript with web components instead of Alpine.js
- **Benefit**: Type safety, better tooling support, more familiar to developers
- **Cost**: Slightly more verbose than Alpine.js for simple state management
- **Justification**: Better long-term maintainability and type safety

### Package Manager Choice
- **Chosen**: Bun over npm/yarn/pnpm
- **Benefit**: 30x faster installs, all-in-one toolkit, minimal dependencies
- **Cost**: Newer ecosystem, potential compatibility issues with some packages
- **Mitigation**: Bun maintains npm compatibility; fallback to npm if needed

### Testing Approach
- **Chosen**: Bun Test + Playwright with comprehensive coverage
- **Benefit**: Built-in testing, real application flow testing, security testing
- **Cost**: Learning curve for Playwright patterns
- **Justification**: Provides excellent test coverage and security validation

### Styling Strategy
- **Chosen**: Tailwind CSS with minimal custom styles
- **Benefit**: Rapid development, consistent design system, small bundle when purged
- **Cost**: Learning curve for utility classes
- **Justification**: Industry standard with excellent performance characteristics

### Build System
- **Chosen**: Bun's built-in bundler over Vite/Webpack
- **Benefit**: No additional dependencies, fast builds, integrated toolchain
- **Cost**: Less mature than Vite, fewer plugins
- **Justification**: Aligns with minimal dependency strategy

### Integration Method
- **Chosen**: Go backend serves static files from root
- **Benefit**: Simple deployment, no separate server, follows design 002
- **Cost**: Less flexibility than separate frontend server
- **Justification**: Matches project's simplicity goals and existing architecture

## Implementation Plan

### Phase 1: Development Environment Setup
1. Set up frontend directory structure
2. Install and configure Bun as package manager
3. Configure Tailwind CSS and TypeScript with minimal dependencies
4. Set up Makefile integration for frontend tasks
5. Create basic HTML structure and TypeScript component skeleton

### Phase 2: Backend Integration & API Connection
1. Integrate Go backend to serve frontend files from root (`/`)
2. Connect to existing API endpoints (`/api/catalog/v1/components`)
3. Create working components list page with real data
4. Add error handling, loading states, and basic responsive design
5. Configure development proxy from Bun to Go backend

### Phase 3: Testing Foundation
1. Set up Bun's built-in test runner for unit tests
2. Configure Playwright with comprehensive test scenarios
3. Create security testing for XSS prevention
4. Write comprehensive unit and e2e test scenarios
5. Add type checking and code quality checks

### Phase 4: Production & CI/CD
1. Optimize build process and bundle size
2. Integrate frontend tasks into existing CI/CD pipeline
3. Add performance monitoring and optimization
4. Complete documentation and deployment guides

## Best Practices References

### Bun
- [Bun Documentation](https://bun.sh/docs)
- [Bun Package Manager](https://bun.sh/docs/cli/install)
- [Bun Test Runner](https://bun.sh/docs/test/runner)
- [Migrating from npm to Bun](https://bun.sh/guides/migrate-from-npm)

### TypeScript
- [TypeScript Documentation](https://www.typescriptlang.org/docs/)
- [TypeScript Best Practices](https://www.typescriptlang.org/docs/handbook/intro.html)
- [Web Components with TypeScript](https://developer.mozilla.org/en-US/docs/Web/Web_Components)

### CSS Design System
- [CSS Custom Properties](https://developer.mozilla.org/en-US/docs/Web/CSS/--*)
- [CSS Design Tokens](https://css-tricks.com/what-are-design-tokens/)
- [Lit CSS-in-JS](https://lit.dev/docs/components/styles/)

### Playwright Testing
- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Page Object Models](https://playwright.dev/docs/pom)
- [Testing with Real Application Flow](https://playwright.dev/docs/locators)

### Package Management Best Practices
- [Bun vs npm Performance Comparison](https://bun.sh/blog/bun-install)
- [Minimizing Dependencies Guide](https://docs.npmjs.com/cli/v7/using-npm/dependency-management)
- [Frontend Build Optimization](https://web.dev/fast/)

This design establishes a solid, minimal-dependency foundation for the web UI that prioritizes speed, simplicity, and maintainability while integrating seamlessly with the existing backend architecture.
