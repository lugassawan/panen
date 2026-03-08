import { describe, expect, it, vi } from "vitest";

vi.mock("../i18n", () => ({
  t: vi.fn((key: string) => {
    const translations: Record<string, string> = {
      "error.ERR_HAS_HOLDINGS": "Portfolio has holdings. Remove them first.",
      "error.ERR_DUPLICATE_MODE": "This mode already exists for this account.",
    };
    return translations[key] ?? key;
  }),
}));

import { formatError } from "./error";

describe("formatError", () => {
  it("translates known error codes", () => {
    expect(formatError("ERR_HAS_HOLDINGS|portfolio has holdings: 2 holding(s) linked")).toBe(
      "Portfolio has holdings. Remove them first.",
    );
  });

  it("translates another known code", () => {
    expect(formatError("ERR_DUPLICATE_MODE|portfolio mode already exists")).toBe(
      "This mode already exists for this account.",
    );
  });

  it("falls back to message portion for unknown codes", () => {
    expect(formatError("ERR_UNKNOWN|something went wrong")).toBe("something went wrong");
  });

  it("returns raw string when no pipe delimiter", () => {
    expect(formatError("plain error message")).toBe("plain error message");
  });

  it("returns raw string when prefix is not ERR_", () => {
    expect(formatError("INFO|not an error")).toBe("INFO|not an error");
  });
});
