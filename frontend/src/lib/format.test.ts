import { describe, expect, it } from "vitest";
import { formatDecimal, formatPercent, formatRupiah } from "./format";

describe("formatRupiah", () => {
  it("formats a typical stock price", () => {
    expect(formatRupiah(9250)).toBe("Rp\u00A09.250");
  });

  it("formats zero", () => {
    expect(formatRupiah(0)).toBe("Rp\u00A00");
  });

  it("formats large values with thousand separators", () => {
    expect(formatRupiah(1500000)).toBe("Rp\u00A01.500.000");
  });

  it("formats decimal values by rounding", () => {
    expect(formatRupiah(9250.75)).toBe("Rp\u00A09.251");
  });
});

describe("formatDecimal", () => {
  it("formats with default 2 digits", () => {
    expect(formatDecimal(1.5678)).toBe("1,57");
  });

  it("formats with custom digits", () => {
    expect(formatDecimal(12.3, 1)).toBe("12,3");
  });

  it("formats zero", () => {
    expect(formatDecimal(0)).toBe("0,00");
  });

  it("formats negative values", () => {
    expect(formatDecimal(-7.89, 2)).toBe("-7,89");
  });
});

describe("formatPercent", () => {
  it("formats a percentage value", () => {
    expect(formatPercent(25.5)).toBe("25,50%");
  });

  it("formats with custom digits", () => {
    expect(formatPercent(33.333, 1)).toBe("33,3%");
  });

  it("formats zero", () => {
    expect(formatPercent(0)).toBe("0,00%");
  });

  it("formats negative percentages", () => {
    expect(formatPercent(-12.5)).toBe("-12,50%");
  });
});
