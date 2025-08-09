import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import "../../src/ui/ui-badge";

describe("ui-badge", () => {
  let el: HTMLElement;

  beforeEach(() => {
    el = document.createElement("ui-badge");
    document.body.appendChild(el);
  });

  afterEach(() => {
    document.body.innerHTML = "";
  });

  function getClasses() {
    return (el.className || "").split(/\s+/);
  }

  test("renders pass variant with green classes and check icon", () => {
    el.setAttribute("status", "pass");
    expect(getClasses()).toContain("bg-green-100");
    expect(getClasses()).toContain("text-green-800");
    expect(el.innerHTML).toContain("svg");
    expect(el.textContent?.trim()).toContain("pass");
  });

  test("renders fail/error/unknown as red", () => {
    for (const s of ["fail", "error", "unknown"]) {
      el.setAttribute("status", s);
      expect(getClasses()).toContain("bg-red-100");
      expect(getClasses()).toContain("text-red-800");
      expect(el.textContent?.trim()).toContain(s);
    }
  });

  test("renders disabled/skipped as yellow", () => {
    for (const s of ["disabled", "skipped"]) {
      el.setAttribute("status", s);
      expect(getClasses()).toContain("bg-yellow-100");
      expect(getClasses()).toContain("text-yellow-800");
    }
  });

  test("renders completed as blue", () => {
    el.setAttribute("status", "completed");
    expect(getClasses()).toContain("bg-blue-100");
    expect(getClasses()).toContain("text-blue-800");
  });

  test("renders default gray for unknown value", () => {
    el.setAttribute("status", "weird");
    expect(getClasses()).toContain("bg-gray-100");
    expect(getClasses()).toContain("text-gray-800");
    expect(el.textContent?.trim()).toContain("weird");
  });

  test("escapes label text", () => {
    el.setAttribute("status", '<img src=x onerror="1">');
    expect(el.innerHTML).not.toContain("<img");
    expect(el.textContent).toContain('<img src=x onerror="1">');
  });

  test("is case-insensitive for classes but preserves label", () => {
    el.setAttribute("status", "PASS");
    expect(getClasses()).toContain("bg-green-100");
    expect(getClasses()).toContain("text-green-800");
    expect(el.textContent?.trim()).toBe("PASS");
  });

  test("renders default when no status attribute provided", () => {
    // No attribute set
    const el2 = document.createElement("ui-badge");
    document.body.appendChild(el2);
    expect(el2.className || "").toContain("bg-gray-100");
    expect(el2.className || "").toContain("text-gray-800");
    expect(el2.textContent?.trim()).toBe("default");
  });
});
