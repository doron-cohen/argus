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
- **Alpine.js**: Lightweight JavaScript framework for reactive UI components
- **Tailwind CSS**: Utility-first CSS framework for minimal styling
- **Playwright**: End-to-end testing framework with mini-flows pattern

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
│   ├── components/
│   │   └── components-list.js
│   ├── styles/
│   │   └── app.css
│   └── app.js
├── tests/
│   ├── unit/
│   │   └── components-list.test.js
│   ├── e2e/
│   │   ├── components-list.spec.js
│   │   └── mini-flows/
│   │       ├── navigation.js
│   │       └── component-actions.js
├── dist/
│   └── index.html
├── package.json
├── tailwind.config.js
├── playwright.config.js
└── bun.lockb
```

### URL Structure
**Backend serves frontend from root**:
- `/` - Components list page (main page)
- `/api/*` - API endpoints (existing)
- Frontend files served from `/` route by Go backend

### Components List Page
**Simple table/list view showing:**
- Component name
- Component ID
- Description (truncated)
- Owner/team
- Last updated timestamp

**Alpine.js Component Structure:**
```javascript
// components-list.js
export default function() {
    return {
        components: [],
        loading: true,
        error: null,
        
        async init() {
            await this.loadComponents()
        },
        
        async loadComponents() {
            try {
                const response = await fetch('/api/catalog/v1/components')
                this.components = await response.json()
            } catch (err) {
                this.error = 'Failed to load components'
            } finally {
                this.loading = false
            }
        }
    }
}
```

### Styling Approach
**Tailwind CSS with minimal custom styles:**
- Use Tailwind utility classes for layout and basic styling
- Minimal custom CSS for component-specific needs
- Focus on readability and usability over aesthetics
- Responsive design for different screen sizes

### Testing Strategy

#### Unit Tests (Bun Test)
**Test Alpine.js components using Bun's built-in test runner:**
```javascript
// components-list.test.js
import { test, expect } from 'bun:test'
import ComponentsList from '../src/components/components-list.js'

test('ComponentsList loads components on init', async () => {
    // Mock fetch and test component behavior
    const mockFetch = jest.fn().mockResolvedValue({
        json: () => Promise.resolve([{ id: '1', name: 'test-component' }])
    })
    global.fetch = mockFetch
    
    const component = ComponentsList()
    await component.init()
    
    expect(component.components).toHaveLength(1)
    expect(component.loading).toBe(false)
})

test('ComponentsList handles error state', async () => {
    global.fetch = jest.fn().mockRejectedValue(new Error('API Error'))
    
    const component = ComponentsList()
    await component.init()
    
    expect(component.error).toBe('Failed to load components')
})
```

#### End-to-End Tests (Playwright + Mini-flows)
**Mini-flows pattern for reusable test components:**
```javascript
// mini-flows/navigation.js
export async function navigateToComponents(page) {
    await page.goto('/')
    await expect(page.locator('h1')).toContainText('Components')
}

export async function waitForComponentsLoad(page) {
    await page.waitForSelector('[data-testid="components-table"]')
    await expect(page.locator('[data-testid="loading"]')).not.toBeVisible()
}

// mini-flows/component-actions.js
export async function searchComponents(page, query) {
    await page.fill('[data-testid="search-input"]', query)
    await page.keyboard.press('Enter')
}

// components-list.spec.js
import { test, expect } from '@playwright/test'
import { navigateToComponents, waitForComponentsLoad } from './mini-flows/navigation.js'
import { searchComponents } from './mini-flows/component-actions.js'

test.describe('Components List Page', () => {
    test('displays components list', async ({ page }) => {
        await navigateToComponents(page)
        await waitForComponentsLoad(page)
        await expect(page.locator('table')).toBeVisible()
    })
    
    test('search functionality works', async ({ page }) => {
        await navigateToComponents(page)
        await waitForComponentsLoad(page)
        await searchComponents(page, 'auth-service')
        await expect(page.locator('tbody tr')).toContainText('auth-service')
    })
})
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
	cd frontend && bun run build

frontend/test:
	cd frontend && bun test

frontend/test-e2e:
	cd frontend && bun run test:e2e

frontend/lint:
	cd frontend && bun run lint

frontend/ci: frontend/install frontend/lint frontend/test frontend/build frontend/test-e2e

# Combined targets
all: backend/gen-all backend/go-mod-tidy frontend/ci

ci: backend/ci frontend/ci
```

**Bun-based development workflow:**
1. **Install**: `make frontend/install` - Fast dependency installation
2. **Development**: `bun dev` - Hot reload with built-in dev server
3. **Testing**: `bun test` - Built-in test runner (no Jest needed)
4. **E2E Testing**: `bun run test:e2e` - Playwright tests
5. **Build**: `bun run build` - Production build with built-in bundler
6. **Lint**: `bun run lint` - Code quality checks

### Integration with Backend
**Go backend serves frontend files:**
- Frontend build output served from `/` route in Go
- API routes remain at `/api/*` (design 002 compliance)
- Development: Bun dev server with API proxy to Go backend
- Production: Go serves static files from `frontend/dist/`

## Tradeoffs

### Package Manager Choice
- **Chosen**: Bun over npm/yarn/pnpm
- **Benefit**: 30x faster installs, all-in-one toolkit, minimal dependencies
- **Cost**: Newer ecosystem, potential compatibility issues with some packages
- **Mitigation**: Bun maintains npm compatibility; fallback to npm if needed

### Framework Choice
- **Chosen**: Alpine.js for simplicity and minimal overhead
- **Benefit**: Lightweight, minimal build complexity, easy to understand
- **Cost**: Less ecosystem support than React/Vue, manual state management
- **Justification**: Perfect for simple component catalog display needs

### Testing Approach
- **Chosen**: Bun Test + Playwright with mini-flows
- **Benefit**: Built-in testing, reusable test patterns, comprehensive coverage
- **Cost**: Learning curve for mini-flows pattern
- **Justification**: Reduces external dependencies while improving test maintainability

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
3. Configure Tailwind CSS and Alpine.js with minimal dependencies
4. Set up Makefile integration for frontend tasks
5. Create basic HTML structure and Alpine.js component skeleton

### Phase 2: Backend Integration & API Connection
1. Integrate Go backend to serve frontend files from root (`/`)
2. Connect to existing API endpoints (`/api/catalog/v1/components`)
3. Create working components list page with real data
4. Add error handling, loading states, and basic responsive design
5. Configure development proxy from Bun to Go backend

### Phase 3: Testing Foundation
1. Set up Bun's built-in test runner for unit tests
2. Configure Playwright with mini-flows pattern
3. Create reusable test helpers and utilities
4. Write comprehensive unit and e2e test scenarios
5. Add linting and code quality checks

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

### Alpine.js
- [Alpine.js Documentation](https://alpinejs.dev/docs/start-here)
- [Alpine.js Best Practices](https://alpinejs.dev/docs/best-practices)
- [Testing Alpine.js Applications](https://alpinejs.dev/docs/advanced/testing)

### Tailwind CSS
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Tailwind CSS Best Practices](https://tailwindcss.com/docs/best-practices)
- [Optimizing for Production](https://tailwindcss.com/docs/optimizing-for-production)

### Playwright Testing
- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Page Object Models](https://playwright.dev/docs/pom)
- [Testing with Mini-flows Pattern](https://playwright.dev/docs/locators)

### Package Management Best Practices
- [Bun vs npm Performance Comparison](https://bun.sh/blog/bun-install)
- [Minimizing Dependencies Guide](https://docs.npmjs.com/cli/v7/using-npm/dependency-management)
- [Frontend Build Optimization](https://web.dev/fast/)

This design establishes a solid, minimal-dependency foundation for the web UI that prioritizes speed, simplicity, and maintainability while integrating seamlessly with the existing backend architecture.
