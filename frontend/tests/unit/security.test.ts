import { describe, test, expect } from "bun:test";
import { escapeHtml } from "../../src/utils.ts";

describe("Security Tests", () => {
  describe("XSS Prevention", () => {
    test("should prevent XSS in component names", () => {
      const maliciousName = "<script>alert('XSS')</script>";
      const escapedName = escapeHtml(maliciousName);
      expect(escapedName).toBe(
        "&lt;script&gt;alert(&#039;XSS&#039;)&lt;/script&gt;",
      );
    });

    test("should prevent XSS in component descriptions", () => {
      const maliciousDescription = "<img src=x onerror=alert('XSS')>";
      const escapedDescription = escapeHtml(maliciousDescription);
      expect(escapedDescription).toBe(
        "&lt;img src=x onerror=alert(&#039;XSS&#039;)&gt;",
      );
    });

    test("should prevent XSS in team names", () => {
      const maliciousTeam = "<script>alert('team')</script>";
      const escapedTeam = escapeHtml(maliciousTeam);
      expect(escapedTeam).toBe(
        "&lt;script&gt;alert(&#039;team&#039;)&lt;/script&gt;",
      );
    });

    test("should prevent XSS in maintainer names", () => {
      const maliciousMaintainer = "<script>alert('maintainer')</script>";
      const escapedMaintainer = escapeHtml(maliciousMaintainer);
      expect(escapedMaintainer).toBe(
        "&lt;script&gt;alert(&#039;maintainer&#039;)&lt;/script&gt;",
      );
    });

    test("should prevent XSS in error messages", () => {
      const maliciousError = "<script>alert('error')</script>";
      const escapedError = escapeHtml(maliciousError);
      expect(escapedError).toBe(
        "&lt;script&gt;alert(&#039;error&#039;)&lt;/script&gt;",
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

  describe("Component Rendering Integration", () => {
    test("should render components with escaped content", () => {
      // Mock component data with malicious content
      const maliciousComponents = [
        {
          id: "xss-test",
          name: "<script>alert('XSS')</script>",
          description: "<img src=x onerror=alert('XSS')>",
          owners: {
            team: "<script>alert('team')</script>",
            maintainers: ["<script>alert('maintainer')</script>"],
          },
        },
      ];

      // Simulate the rendering process
      const renderComponent = (comp: any) => {
        return `
          <div data-testid="component-name">${escapeHtml(comp.name)}</div>
          <div data-testid="component-description">${escapeHtml(
            comp.description,
          )}</div>
          <div data-testid="component-team">${escapeHtml(
            comp.owners?.team || "",
          )}</div>
          <div data-testid="component-maintainers">${escapeHtml(
            comp.owners?.maintainers?.join(", ") || "",
          )}</div>
        `;
      };

      const rendered = renderComponent(maliciousComponents[0]);

      // Verify that all malicious content is escaped
      expect(rendered).toContain(
        "&lt;script&gt;alert(&#039;XSS&#039;)&lt;/script&gt;",
      );
      expect(rendered).toContain(
        "&lt;img src=x onerror=alert(&#039;XSS&#039;)&gt;",
      );
      expect(rendered).toContain(
        "&lt;script&gt;alert(&#039;team&#039;)&lt;/script&gt;",
      );
      expect(rendered).toContain(
        "&lt;script&gt;alert(&#039;maintainer&#039;)&lt;/script&gt;",
      );

      // Verify no unescaped malicious HTML is present
      expect(rendered).not.toContain("<script>alert('XSS')</script>");
      expect(rendered).not.toContain("<img src=x onerror=alert('XSS')>");
      expect(rendered).not.toContain("<script>alert('team')</script>");
      expect(rendered).not.toContain("<script>alert('maintainer')</script>");
    });
  });
});
