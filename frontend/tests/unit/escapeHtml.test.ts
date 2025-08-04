import { describe, test, expect } from "bun:test";
import { escapeHtml } from "../../src/utils.ts";

describe("escapeHtml function", () => {
  test("should escape basic HTML characters", () => {
    const input = "<script>alert('XSS')</script>";
    const expected = "&lt;script&gt;alert(&#039;XSS&#039;)&lt;/script&gt;";
    expect(escapeHtml(input)).toBe(expected);
  });

  test("should escape all HTML entities", () => {
    const input = "& < > \" '";
    const expected = "&amp; &lt; &gt; &quot; &#039;";
    expect(escapeHtml(input)).toBe(expected);
  });

  test("should handle normal text without escaping", () => {
    const input = "Normal text without special characters";
    expect(escapeHtml(input)).toBe(input);
  });

  test("should handle empty string", () => {
    expect(escapeHtml("")).toBe("");
  });

  test("should handle null and undefined", () => {
    expect(escapeHtml(null)).toBe("null");
    expect(escapeHtml(undefined)).toBe("undefined");
  });

  test("should escape img tags", () => {
    const input = "<img src=x onerror=alert('XSS')>";
    const expected = "&lt;img src=x onerror=alert(&#039;XSS&#039;)&gt;";
    expect(escapeHtml(input)).toBe(expected);
  });

  test("should escape complex HTML", () => {
    const input = '<div class="test" onclick="alert(\'XSS\')">Content</div>';
    const expected =
      "&lt;div class=&quot;test&quot; onclick=&quot;alert(&#039;XSS&#039;)&quot;&gt;Content&lt;/div&gt;";
    expect(escapeHtml(input)).toBe(expected);
  });

  test("should handle already escaped content", () => {
    const input = "&amp; &lt; &gt; &quot; &#039;";
    const expected = "&amp;amp; &amp;lt; &amp;gt; &amp;quot; &amp;#039;";
    expect(escapeHtml(input)).toBe(expected);
  });

  test("should handle mixed content", () => {
    const input = "Normal text <script>alert('XSS')</script> more text";
    const expected =
      "Normal text &lt;script&gt;alert(&#039;XSS&#039;)&lt;/script&gt; more text";
    expect(escapeHtml(input)).toBe(expected);
  });

  test("should handle special characters in different contexts", () => {
    const testCases = [
      { input: "User's name", expected: "User&#039;s name" },
      { input: 'User "quoted" text', expected: "User &quot;quoted&quot; text" },
      { input: "A & B", expected: "A &amp; B" },
      { input: "x < y > z", expected: "x &lt; y &gt; z" },
    ];

    testCases.forEach(({ input, expected }) => {
      expect(escapeHtml(input)).toBe(expected);
    });
  });

  test("should handle unicode characters", () => {
    const input = "Hello 世界 <script>alert('XSS')</script>";
    const expected =
      "Hello 世界 &lt;script&gt;alert(&#039;XSS&#039;)&lt;/script&gt;";
    expect(escapeHtml(input)).toBe(expected);
  });

  test("should handle very long strings", () => {
    const longString =
      "<script>".repeat(1000) + "alert('XSS')" + "</script>".repeat(1000);
    const escaped = escapeHtml(longString);

    // Should not contain any unescaped script tags
    expect(escaped).not.toContain("<script>");
    expect(escaped).not.toContain("</script>");

    // Should contain escaped versions
    expect(escaped).toContain("&lt;script&gt;");
    expect(escaped).toContain("&lt;/script&gt;");
  });

  test("should handle all XSS vectors", () => {
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
