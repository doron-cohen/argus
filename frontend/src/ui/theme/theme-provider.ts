export type Theme = "light" | "dark";

export function setTheme(theme: Theme): void {
  document.documentElement.setAttribute("data-theme", theme);
}

// Initialize default theme once
if (!document.documentElement.getAttribute("data-theme")) {
  setTheme("light");
}
