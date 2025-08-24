## ADR: CSS Tokens and Theme Groundwork

### Context
We are migrating from Tailwind-first styling to Lit components that consume semantic CSS custom properties. This enables theming (light/dark) that cascades through shadow DOM via custom properties.

### Decision
- Introduce core and semantic CSS tokens loaded once globally via `src/styles.css`.
- Provide `data-theme` attribute on `:root` to switch themes.
- Keep Tailwind for layout and utility usage during the transition.
- First adopter: `ui-badge` component uses semantic tokens with fallbacks.

### Implementation
- Tokens: `frontend/src/ui/tokens/{core.css, palette.css, semantic.css, themes/{light.css,dark.css}}`
- Bootstrap: import tokens via `src/styles.css` and set theme via `src/ui/theme/theme-provider.ts`.
- Component migration pattern: replace hard-coded colors with semantic variables.

### Consequences
- Components remain functional without tokens thanks to fallbacks.
- Subsequent components can migrate incrementally.


