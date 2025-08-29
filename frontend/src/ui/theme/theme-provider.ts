import { themeStore } from "./theme-store";

export type Theme = "light" | "dark";

export function setTheme(theme: Theme): void {
  document.documentElement.setAttribute("data-theme", theme);
  localStorage.setItem("theme", theme);
  themeStore.set(theme);
}

export function getStoredTheme(): Theme | null {
  return localStorage.getItem("theme") as Theme | null;
}

export function getSystemTheme(): Theme {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

export function initializeTheme(): void {
  const stored = getStoredTheme();
  const theme = stored || getSystemTheme();
  setTheme(theme);
}

// Initialize theme once
if (!document.documentElement.getAttribute("data-theme")) {
  initializeTheme();
}

// Listen for system theme changes when no stored preference exists
if (!getStoredTheme()) {
  const mq = window.matchMedia("(prefers-color-scheme: dark)");
  mq.addEventListener("change", (e) => setTheme(e.matches ? "dark" : "light"));
}

// Cross-tab sync
window.addEventListener("storage", (e) => {
  if (e.key === "theme" && e.newValue) setTheme(e.newValue as Theme);
});
