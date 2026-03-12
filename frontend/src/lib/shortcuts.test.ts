import { describe, expect, it, vi } from "vitest";
import { handleGlobalShortcut, type ShortcutHandlers } from "./shortcuts";

function makeHandlers(overrides: Partial<ShortcutHandlers> = {}): ShortcutHandlers {
  return {
    onNavigate: vi.fn(),
    onToggleCommandPalette: vi.fn(),
    onToggleHelp: vi.fn(),
    ...overrides,
  };
}

function fire(key: string, opts: Partial<KeyboardEventInit> = {}): KeyboardEvent {
  const e = new KeyboardEvent("keydown", { key, bubbles: true, ...opts });
  vi.spyOn(e, "preventDefault");
  return e;
}

describe("handleGlobalShortcut", () => {
  it("Cmd+K toggles command palette", () => {
    const h = makeHandlers();
    const e = fire("k", { metaKey: true });
    handleGlobalShortcut(e, h);
    expect(h.onToggleCommandPalette).toHaveBeenCalled();
    expect(e.preventDefault).toHaveBeenCalled();
  });

  it("Cmd+1 navigates to lookup", () => {
    const h = makeHandlers();
    const e = fire("1", { metaKey: true });
    handleGlobalShortcut(e, h);
    expect(h.onNavigate).toHaveBeenCalledWith("lookup");
  });

  it("Shift+? toggles help", () => {
    const h = makeHandlers();
    const e = fire("?", { shiftKey: true });
    handleGlobalShortcut(e, h);
    expect(h.onToggleHelp).toHaveBeenCalled();
  });

  it("/ opens command palette", () => {
    const h = makeHandlers();
    const e = fire("/");
    handleGlobalShortcut(e, h);
    expect(h.onToggleCommandPalette).toHaveBeenCalled();
  });

  it("blocks shortcuts when input is focused except Cmd+K and Escape", () => {
    const input = document.createElement("input");
    document.body.appendChild(input);

    const h = makeHandlers();

    // "/" should be blocked in input
    const slashEvent = new KeyboardEvent("keydown", { key: "/", bubbles: true });
    Object.defineProperty(slashEvent, "target", { value: input });
    handleGlobalShortcut(slashEvent, h);
    expect(h.onToggleCommandPalette).not.toHaveBeenCalled();

    // Cmd+K should still work in input
    const cmdKEvent = new KeyboardEvent("keydown", { key: "k", metaKey: true, bubbles: true });
    Object.defineProperty(cmdKEvent, "target", { value: input });
    vi.spyOn(cmdKEvent, "preventDefault");
    handleGlobalShortcut(cmdKEvent, h);
    expect(h.onToggleCommandPalette).toHaveBeenCalled();

    document.body.removeChild(input);
  });
});
