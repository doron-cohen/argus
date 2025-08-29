# Theme Customization Guide

This document explains how to customize and extend the Argus theme system, which provides robust theming, accessibility, and design token management.

## Overview

The Argus theme system is built with CSS custom properties (CSS variables) and follows a design token architecture. It supports light and dark themes out of the box, with comprehensive accessibility features including WCAG AA compliance.

## Architecture

### Core Principles

1. **Design Tokens First**: All colors, spacing, typography, and other design values are defined as CSS custom properties
2. **Semantic Naming**: Colors and other tokens use semantic names (e.g., `--color-bg`, `--color-fg`) rather than descriptive names
3. **Theme Overrides**: Themes override semantic tokens to achieve different visual appearances
4. **Accessibility Built-in**: Focus management, contrast ratios, and motion preferences are handled automatically

### File Structure

```
src/ui/
├── tokens/
│   ├── core.css          # Core design tokens (spacing, typography, shadows)
│   ├── palette.css       # Color palette (raw colors)
│   ├── semantic.css      # Semantic color tokens (default mappings)
│   ├── themes/
│   │   ├── light.css     # Light theme overrides
│   │   └── dark.css      # Dark theme overrides
│   └── reset.css         # Base styles with accessibility features
├── theme/
│   └── theme-provider.ts # Theme management utilities
└── utilities.css         # Utility classes
```

## Using the Theme System

### Basic Theme Switching

```typescript
import { setTheme, getSystemTheme } from './src/ui/theme/theme-provider';

// Set a specific theme
setTheme('dark');

// Get system preference
const systemTheme = getSystemTheme();

// Initialize with stored preference or system default
initializeTheme();
```

### CSS Custom Properties

All design tokens are available as CSS custom properties:

```css
.my-component {
  /* Core tokens */
  padding: var(--space-4);
  font-size: var(--font-size-md);
  border-radius: var(--radius-2);

  /* Semantic color tokens */
  background-color: var(--color-bg);
  color: var(--color-fg);
  border: 1px solid var(--color-border);

  /* Component-specific tokens */
  transition: all var(--transition-duration, 0.15s) ease;
}
```

## Available Token Categories

### Spacing Scale
```css
--space-0: 0rem;
--space-0-5: 0.125rem;
--space-1: 0.25rem;
/* ... up to --space-12: 3rem */
```

### Typography Scale
```css
--font-size-xs: 0.75rem;
--font-size-sm: 0.875rem;
--font-size-md: 1rem;
/* ... up to --font-size-4xl: 2.25rem */

--font-weight-regular: 400;
--font-weight-medium: 500;
/* ... up to --font-weight-bold: 700 */
```

### Color Palette
```css
/* Grays */
--palette-gray-50: rgb(249 250 251);
--palette-gray-100: rgb(243 244 246);
/* ... up to --palette-gray-900: rgb(17 24 39) */

/* Semantic colors */
--palette-green-100: rgb(220 252 231);
--palette-green-600: rgb(22 163 74);
/* ... and more for red, yellow, blue */
```

### Semantic Colors
```css
/* Base */
--color-bg: var(--palette-white);
--color-fg: var(--palette-gray-900);
--color-border: var(--palette-gray-200);

/* Status */
--color-success-bg: var(--palette-green-100);
--color-success-fg: var(--palette-green-600);

/* Interactive */
--color-info-bg: var(--palette-blue-100);
--color-info-fg: var(--palette-blue-700);
```

## Creating Custom Themes

### Method 1: Override Existing Themes

Create a new theme file in `src/ui/tokens/themes/`:

```css
/* src/ui/tokens/themes/custom.css */
:root[data-theme="custom"] {
  /* Override semantic colors */
  --color-bg: var(--palette-gray-50);
  --color-fg: var(--palette-gray-900);
  --color-border: var(--palette-gray-300);

  /* Custom accent colors */
  --color-accent-bg: var(--palette-purple-100);
  --color-accent-fg: var(--palette-purple-700);
}
```

Import it in your main CSS file:

```css
@import './src/ui/tokens/themes/custom.css';
```

### Method 2: Runtime Theme Customization

Use CSS custom properties to override themes dynamically:

```typescript
// Apply custom theme properties
document.documentElement.style.setProperty('--color-primary', '#your-color');
document.documentElement.style.setProperty('--space-md', '1.5rem');
```

### Method 3: Component-Level Customization

Override tokens for specific components:

```css
.my-special-button {
  /* Override semantic tokens locally */
  --color-info-bg: var(--palette-purple-100);
  --color-info-fg: var(--palette-purple-700);
}
```

## Accessibility Features

### Focus Management

The theme system includes comprehensive focus management:

```css
/* Automatic focus rings for all elements */
*:focus-visible {
  outline: 2px solid var(--color-info-fg);
  outline-offset: 2px;
}

/* Theme-aware focus colors */
:root[data-theme="dark"] *:focus-visible {
  outline-color: var(--palette-blue-300);
}
```

### Motion Preferences

Respects `prefers-reduced-motion`:

```css
/* Disable animations when user prefers reduced motion */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

### High Contrast Support

Enhanced focus for high contrast mode:

```css
@media (prefers-contrast: high) {
  *:focus-visible {
    outline-width: 3px;
    outline-style: solid;
  }
}
```

## Utility Classes

The system provides utility classes for common patterns:

```css
/* Transitions */
.transition-safe { transition: all 0.15s ease; }
.transition-colors { transition: color 0.15s ease, background-color 0.15s ease; }

/* Focus utilities */
.focus-ring { outline: 2px solid var(--color-info-fg); }
.focus-ring-inset { outline: 2px solid var(--color-info-fg); outline-offset: -2px; }
```

## Best Practices

### 1. Always Use Semantic Tokens

```css
/* ✅ Good */
.my-component {
  background-color: var(--color-bg);
  color: var(--color-fg);
}

/* ❌ Avoid */
.my-component {
  background-color: white;
  color: black;
}
```

### 2. Provide Fallbacks

```css
.my-component {
  padding: var(--space-4, 1rem);
  font-size: var(--font-size-md, 1rem);
}
```

### 3. Test in Both Themes

Always test your components in both light and dark themes to ensure proper contrast and visibility.

### 4. Use Utility Classes for Consistency

```css
.my-component {
  @apply transition-colors;
  @apply focus-ring;
}
```

## Advanced Customization

### Creating a Brand Theme

```css
/* Brand colors */
:root {
  --brand-primary: #your-brand-color;
  --brand-secondary: #your-secondary-color;
}

/* Apply to semantic tokens */
:root[data-theme="light"] {
  --color-info-bg: var(--brand-primary);
  --color-info-fg: white;
}

:root[data-theme="dark"] {
  --color-info-bg: var(--brand-primary);
  --color-info-fg: white;
}
```

### Dynamic Theme Switching

```typescript
import { setTheme } from './theme-provider';

// Toggle theme
function toggleTheme() {
  const currentTheme = document.documentElement.getAttribute('data-theme');
  const newTheme = currentTheme === 'light' ? 'dark' : 'light';
  setTheme(newTheme);
}
```

### CSS-in-JS Integration

```typescript
const themeStyles = css`
  background-color: var(--color-bg);
  color: var(--color-fg);
  padding: var(--space-4);
`;
```

## Testing

The theme system includes comprehensive tests:

- **Visual regression tests**: Screenshots for theme consistency
- **Contrast validation**: WCAG AA compliance testing
- **Accessibility audits**: axe-core integration
- **Theme switching**: Ensures smooth transitions

## Migration from Tailwind

### Before (Tailwind)
```html
<div class="bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 p-4 rounded-md">
  Content
</div>
```

### After (Design Tokens)
```html
<div class="u-bg-primary u-text-primary u-space-4 u-radius-2">
  Content
</div>
```

```css
.u-bg-primary { background-color: var(--color-bg); }
.u-text-primary { color: var(--color-fg); }
.u-space-4 { padding: var(--space-4); }
.u-radius-2 { border-radius: var(--radius-2); }
```

This approach provides:
- Consistent theming across light/dark modes
- Better maintainability
- Improved accessibility
- Smaller bundle size (no utility classes)
- Better performance (fewer CSS rules)
