import { render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

vi.mock("../../../wailsjs/go/backend/App", () => ({
  ListBrokerConfigs: vi.fn(),
  CreateBrokerageAccount: vi.fn(),
  CreatePortfolio: vi.fn(),
}));

import PortfolioOnboarding from "./PortfolioOnboarding.svelte";

describe("PortfolioOnboarding", () => {
  const defaultProps = {
    brokerConfigs: [{ id: "bc1", name: "Stockbit", code: "STOCKBIT" }],
    brokerageAcctId: null,
    showPortfolioForm: false,
    onBrokerageCreated: vi.fn(),
    onPortfolioCreated: vi.fn(),
  };

  it("renders brokerage setup heading when showPortfolioForm is false", () => {
    render(PortfolioOnboarding, { props: defaultProps });
    expect(screen.getByText("Set Up Your Brokerage")).toBeInTheDocument();
  });

  it("renders portfolio creation heading when showPortfolioForm is true", () => {
    render(PortfolioOnboarding, {
      props: { ...defaultProps, showPortfolioForm: true, brokerageAcctId: "b1" },
    });
    expect(screen.getByText("Create Your Portfolio")).toBeInTheDocument();
  });
});
