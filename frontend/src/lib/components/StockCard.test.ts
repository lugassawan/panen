import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import StockCardWrapper from "./__tests__/StockCardWrapper.svelte";

describe("StockCard", () => {
  it("renders ticker and name", () => {
    render(StockCardWrapper, {
      props: { ticker: "BBCA", name: "Bank Central Asia" },
    });
    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText("Bank Central Asia")).toBeInTheDocument();
  });

  it("renders formatted price", () => {
    render(StockCardWrapper, { props: { price: 9500 } });
    expect(screen.getByText(/9[.,]500/)).toBeInTheDocument();
  });

  it("shows positive change with + prefix", () => {
    render(StockCardWrapper, {
      props: { change: 100, changePercent: 1.06 },
    });
    expect(screen.getByText(/\+.*1\.06%/)).toBeInTheDocument();
  });

  it("shows negative change without + prefix", () => {
    render(StockCardWrapper, {
      props: { change: -200, changePercent: -2.1 },
    });
    expect(screen.getByText(/-2\.10%/)).toBeInTheDocument();
  });

  it("renders metrics when provided", () => {
    render(StockCardWrapper, {
      props: {
        metrics: [
          { label: "ROE", value: "12.5%" },
          { label: "PER", value: "18.0" },
        ],
      },
    });
    expect(screen.getByText("ROE")).toBeInTheDocument();
    expect(screen.getByText("12.5%")).toBeInTheDocument();
    expect(screen.getByText("PER")).toBeInTheDocument();
  });

  it("renders value mode badge by default", () => {
    render(StockCardWrapper);
    expect(screen.getByText(/Value/)).toBeInTheDocument();
  });

  it("renders dividend mode badge", () => {
    render(StockCardWrapper, { props: { mode: "dividend" } });
    expect(screen.getByText(/Dividend/)).toBeInTheDocument();
  });
});
