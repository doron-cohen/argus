// HTML escaping function to prevent XSS attacks
export function escapeHtml(unsafe: string | null | undefined): string {
  if (unsafe == null) return String(unsafe);
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/\"/g, "&quot;")
    .replace(/'/g, "&#039;");
}
# Test comment
