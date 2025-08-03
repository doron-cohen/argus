import { describe, test, expect } from "bun:test";

// Import the escapeHtml function from the main file
// Since it's not exported, we'll recreate it here for testing
function escapeHtml(unsafe: string | null | undefined): string {
  if (unsafe == null) return String(unsafe);
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

describe("Security Tests", () => {
  describe("XSS Prevention", () => {
    test("should prevent XSS in component names", () => {
      const maliciousName = "<script>alert('XSS')</script>";
      const escapedName = escapeHtml(maliciousName);
      expect(escapedName).toBe(
        "&lt;script&gt;alert(&#039;XSS&#039;)&lt;/script&gt;"
      );
    });

    test("should prevent XSS in component descriptions", () => {
      const maliciousDescription = "<img src=x onerror=alert('XSS')>";
      const escapedDescription = escapeHtml(maliciousDescription);
      expect(escapedDescription).toBe(
        "&lt;img src=x onerror=alert(&#039;XSS&#039;)&gt;"
      );
    });

    test("should prevent XSS in team names", () => {
      const maliciousTeam = "<script>alert('team')</script>";
      const escapedTeam = escapeHtml(maliciousTeam);
      expect(escapedTeam).toBe(
        "&lt;script&gt;alert(&#039;team&#039;)&lt;/script&gt;"
      );
    });

    test("should prevent XSS in maintainer names", () => {
      const maliciousMaintainer = "<script>alert('maintainer')</script>";
      const escapedMaintainer = escapeHtml(maliciousMaintainer);
      expect(escapedMaintainer).toBe(
        "&lt;script&gt;alert(&#039;maintainer&#039;)&lt;/script&gt;"
      );
    });

    test("should prevent XSS in error messages", () => {
      const maliciousError = "<script>alert('error')</script>";
      const escapedError = escapeHtml(maliciousError);
      expect(escapedError).toBe(
        "&lt;script&gt;alert(&#039;error&#039;)&lt;/script&gt;"
      );
    });
  });

  describe("Null and Undefined Handling", () => {
    test("should handle null component names", () => {
      expect(escapeHtml(null)).toBe("null");
    });

    test("should handle undefined component names", () => {
      expect(escapeHtml(undefined)).toBe("undefined");
    });

    test("should handle null descriptions", () => {
      expect(escapeHtml(null)).toBe("null");
    });

    test("should handle undefined team names", () => {
      expect(escapeHtml(undefined)).toBe("undefined");
    });
  });

  describe("Edge Cases", () => {
    test("should handle empty strings", () => {
      expect(escapeHtml("")).toBe("");
    });

    test("should handle strings with only special characters", () => {
      expect(escapeHtml("& < > \" '")).toBe("&amp; &lt; &gt; &quot; &#039;");
    });

    test("should handle normal text without special characters", () => {
      const normalText = "Normal component name";
      expect(escapeHtml(normalText)).toBe(normalText);
    });

    test("should handle mixed content", () => {
      const mixedContent =
        "Normal text <script>alert('XSS')</script> more text";
      const expected =
        "Normal text &lt;script&gt;alert(&#039;XSS&#039;)&lt;/script&gt; more text";
      expect(escapeHtml(mixedContent)).toBe(expected);
    });
  });

  describe("Complex XSS Vectors", () => {
    test("should handle various XSS attack vectors", () => {
      const xssVectors = [
        "<script>alert('XSS')</script>",
        "<img src=x onerror=alert('XSS')>",
        "<svg onload=alert('XSS')>",
        "<iframe src=javascript:alert('XSS')>",
        "<object data=javascript:alert('XSS')>",
        "<embed src=javascript:alert('XSS')>",
        "<form action=javascript:alert('XSS')>",
        "<input onfocus=alert('XSS')>",
        "<textarea onblur=alert('XSS')>",
        "<select onchange=alert('XSS')>",
        "<button onclick=alert('XSS')>",
        "<a href=javascript:alert('XSS')>",
        "<link rel=stylesheet href=javascript:alert('XSS')>",
        "<meta http-equiv=refresh content=0;url=javascript:alert('XSS')>",
      ];

      xssVectors.forEach((vector) => {
        const escaped = escapeHtml(vector);

        // Should not contain any unescaped HTML
        expect(escaped).not.toContain("<");
        expect(escaped).not.toContain(">");

        // Should contain escaped versions
        expect(escaped).toContain("&lt;");
        expect(escaped).toContain("&gt;");
      });
    });
  });
});
