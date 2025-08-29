# Argus Frontend

A modern web application built with Lit, TypeScript, and a comprehensive design system that supports theming, accessibility, and component reusability.

## Quick Start

### Install Dependencies

```bash
bun install
```

### Development

```bash
# Start development server
bun run dev

# Generate API client
bun run generate:api
```

### Testing

```bash
# Run all tests
bun run test:all

# Run specific test suites
bun run test:unit        # Unit tests
bun run test:e2e         # End-to-end tests
bun run test:ui          # UI component tests
```

### Building

```bash
# Development build
bun run build:dev

# Production build
bun run build:prod
```

## Architecture

This frontend is built with:

- **Lit**: Web Components framework for reusable UI components
- **TypeScript**: Type-safe development
- **Vite**: Fast build tool and dev server
- **Playwright**: End-to-end testing
- **Design System**: Comprehensive theming and component library

## Design System

The application uses a robust design system with:

- **Design Tokens**: CSS custom properties for consistent theming
- **Light/Dark Themes**: Automatic theme switching with system preference detection
- **Accessibility**: WCAG 2.1 AA compliant with focus management and contrast validation
- **Motion Preferences**: Respects `prefers-reduced-motion` for inclusive design

### Theme Customization

For detailed information about customizing themes, theming architecture, and design tokens, see:

- [Theme Customization Guide](../docs/designs/theme-customization.md)
- [Design Token Architecture](../docs/designs/006-css-tokens-and-theme.md)

### Key Features

- **🎨 Theme System**: Light/dark mode with CSS custom properties
- **♿ Accessibility**: WCAG AA compliance, focus management, screen reader support
- **⚡ Performance**: Optimized bundling with Vite, tree-shaking
- **🧪 Testing**: Comprehensive test coverage with visual regression tests
- **📱 Responsive**: Mobile-first design with fluid typography
- **🔧 Developer Experience**: Hot reload, TypeScript, comprehensive tooling

## Development Workflow

### Component Development

1. Create components in `src/ui/components/`
2. Use design tokens for consistent styling
3. Include accessibility features (focus management, ARIA labels)
4. Add unit tests in `src/ui/components/*.test.ts`
5. Test with Playwright for integration

### Theming

Components should use semantic design tokens:

```css
.my-component {
  background-color: var(--color-bg);
  color: var(--color-fg);
  padding: var(--space-4);
  border-radius: var(--radius-2);
}
```

### Accessibility

All components include accessibility features:

- Focus-visible outlines for keyboard navigation
- Proper ARIA labels and roles
- Contrast ratios meeting WCAG AA standards
- Motion preferences respected

## Scripts Reference

| Script | Description |
|--------|-------------|
| `bun run dev` | Start development server |
| `bun run build` | Production build |
| `bun run test:all` | Run all test suites |
| `bun run test:e2e` | End-to-end tests with Playwright |
| `bun run test:unit` | Unit tests with Bun |
| `bun run test:ui` | UI component tests |
| `bun run generate:api` | Generate API client from OpenAPI spec |
| `bun run format` | Format code with Prettier |
| `bun run type-check` | TypeScript type checking |

## Contributing

See the main [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines and coding standards.

## License

See [LICENSE](../LICENSE) for details.
