import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import App from "./App.svelte";

vi.mock("../wailsjs/go/app/App", () => ({
  Greet: vi.fn((name: string) => Promise.resolve(`Hello ${name}, welcome to Panen!`)),
}));

describe("App", () => {
  it("renders the heading", () => {
    render(App);
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent("Panen");
  });

  it("renders the description", () => {
    render(App);
    expect(screen.getByText("Desktop decision engine for IDX investors")).toBeInTheDocument();
  });

  it("renders an input and a greet button", () => {
    render(App);
    expect(screen.getByPlaceholderText("Enter your name")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Greet" })).toBeInTheDocument();
  });

  it("displays a greeting after clicking Greet", async () => {
    const user = userEvent.setup();
    render(App);

    const input = screen.getByPlaceholderText("Enter your name");
    await user.type(input, "Alice");
    await user.click(screen.getByRole("button", { name: "Greet" }));

    expect(await screen.findByText("Hello Alice, welcome to Panen!")).toBeInTheDocument();
  });
});
