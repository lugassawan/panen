# Panen Design System

> **Stack:** Svelte 5 (runes) · Tailwind CSS v4 (CSS-first) · Wails v2
> **Version:** 1.0 — March 2026

---

## Quick Start

```
frontend/src/
├── app.css                         ← Design tokens + Tailwind config (source of truth)
├── lib/
│   ├── stores/
│   │   ├── theme.svelte.ts         ← Light/dark/system theme (Svelte 5 runes)
│   │   └── mode.svelte.ts          ← Value/dividend mode (Svelte 5 runes)
│   └── components/
│       ├── Button.svelte
│       ├── Badge.svelte
│       ├── Alert.svelte
│       ├── ModeTabs.svelte
│       ├── ThemeToggle.svelte
│       └── StockCard.svelte
```

All design tokens live in `frontend/src/app.css`. No `tailwind.config.js` — Tailwind v4 reads everything from CSS via the `@theme` directive.

---

## 1. Architecture Decisions

### Why CSS-first (Tailwind v4)

Tailwind v4 replaces the JS config file with `@theme` blocks in CSS. Tokens defined in `@theme` automatically generate utility classes:

```css
@theme {
  --color-green-700: #1b6b4a;  /* → bg-green-700, text-green-700, border-green-700 */
  --font-display: "Plus Jakarta Sans", sans-serif;  /* → font-display */
  --radius-lg: 12px;  /* → rounded-lg */
}
```

### Why Runes (Svelte 5)

Svelte 5 replaces `writable`/`derived` stores with runes (`$state`, `$derived`, `$effect`). Stores use the "store as object" pattern — `$state` inside a closure returned as a plain object with getters:

```ts
function createThemeStore() {
  let preference = $state<ThemePreference>("system");
  let resolved = $state<ResolvedTheme>("light");
  return {
    get current(): ResolvedTheme { return resolved; },
    get preference(): ThemePreference { return preference; },
    set(pref: ThemePreference) { /* updates both states */ },
    toggle() { /* cycles light → dark → system */ },
  };
}
export const theme = createThemeStore();
```

### Theme Strategy — CSS Variables + Class Toggle

- Static brand colors (`green-*`, `gold-*`) → defined in `@theme`, same in both themes
- Semantic colors (`bg-primary`, `text-secondary`) → defined as CSS variables in `:root` and `.dark`, referenced from `@theme` via `var()`
- Theme switch → toggles `class="dark"` on `<html>`, CSS variables update, all utilities follow automatically

### Wails Consideration

This is a Wails v2 desktop app running in a native webview. There is no SvelteKit, no server-side rendering, and no `$app/environment`. The app is fully client-side. Theme preference is persisted to `localStorage` (safe in Wails webview). All other persistence goes through the Go backend.

---

## 2. Color System

### Brand Colors (same in light and dark)

| Token | Hex | Usage |
|---|---|---|
| `green-700` | `#1b6b4a` | Primary accent, Value Mode, buttons, links |
| `green-800` | `#0f4a32` | Hover states |
| `green-900` | `#0a3524` | App icon background, deep backgrounds |
| `green-100` | `#e6f5ec` | Active nav background (light), subtle tints |
| `gold-500` | `#d4a12a` | Dividend Mode accent, harvest highlights |
| `gold-100` | `#fbf4dc` | Dividend badge background |

Full scales are defined in `app.css`: `green-50` through `green-900`, `gold-50` through `gold-700`.

### Semantic Colors (adapt to theme)

| Token | Light | Dark | Usage |
|---|---|---|---|
| `bg-primary` | `#fefcf7` | `#111112` | Page canvas |
| `bg-secondary` | `#f5f1eb` | `#1a1a1c` | Sidebar, grouped sections |
| `bg-tertiary` | `#ede8df` | `#232326` | Hover backgrounds |
| `bg-elevated` | `#ffffff` | `#1e1e21` | Cards, modals, dropdowns |
| `bg-sunken` | `#f0ebe3` | `#0d0d0e` | Inset panels |
| `text-primary` | `#1a1a1a` | `#eeecea` | Headings, body text |
| `text-secondary` | `#4b5060` | `#a8a5a0` | Supporting text |
| `text-tertiary` | `#6b7280` | `#7a7670` | Labels, descriptions |
| `text-muted` | `#9ca3af` | `#5a5652` | Metric labels |
| `border-default` | `#e0dbd2` | `#2e2e32` | Card borders, dividers |
| `border-subtle` | `#ede8df` | `#252528` | Light separators |
| `positive` | `#1b7d4e` | `#4ade80` | Gains, success states |
| `negative` | `#c4342d` | `#f87171` | Losses, errors |
| `warning` | `#c48a15` | `#fbbf24` | Caution states |

### Mode Colors

Each investment mode tints contextual elements (active nav, badges, tab indicators):

| Mode | Accent | Light bg | Container utility |
|---|---|---|---|
| Value | `green-700` | `green-100` | `mode-value` |
| Dividend | `gold-500` | `gold-100` | `mode-dividend` |

Apply `mode-value` or `mode-dividend` to a parent container. Children reference `var(--mode-accent)` and `var(--mode-accent-light)` directly.

```svelte
<div class={mode.containerClass}>
  <!-- mode.containerClass is "mode-value" or "mode-dividend" -->
  <span style:color="var(--mode-accent)">Active</span>
</div>
```

---

## 3. Typography

### Font Stack

| Token | Font | Usage | Weights |
|---|---|---|---|
| `font-display` | Plus Jakarta Sans | Headings, tickers, brand | 500–800 |
| `font-body` | DM Sans | Body, labels, UI controls | 400–700 |
| `font-mono` | DM Mono | Prices, percentages, ratios | 400–500 |

All three fonts are self-hosted as WOFF2 files in `frontend/src/assets/fonts/` and declared via `@font-face` in `app.css`. No external requests to Google Fonts. The default `html` element uses `font-body`.

### Type Scale

| Class | Size | Usage |
|---|---|---|
| `text-4xl` | 36px | Page titles |
| `text-3xl` | 30px | Section headers |
| `text-2xl` | 24px | Card titles, dialogs |
| `text-xl` | 20px | Subsection titles |
| `text-lg` | 17px | Stock headers, list titles |
| `text-md` | 15px | Important body text |
| `text-base` | 14px | Default body, form fields |
| `text-sm` | 13px | Labels, captions |
| `text-xs` | 11px | Metric labels (use with `uppercase tracking-wide`) |

### Financial Numbers

Always use `font-mono` for financial values. Right-align in tables.

```svelte
<!-- Price with change -->
<span class="font-mono text-xl font-semibold">Rp 4,250</span>
<span class="font-mono text-sm text-profit">+125 (+3.03%)</span>

<!-- Metric label + value -->
<div class="text-xs font-semibold uppercase tracking-wide text-text-muted">PBV</div>
<div class="font-mono text-sm">1.28×</div>
```

---

## 4. Spacing, Radius & Shadows

### Spacing (4px base)

| Class | Value | Common usage |
|---|---|---|
| `p-1` / `gap-1` | 4px | Tight inner gaps |
| `p-2` / `gap-2` | 8px | Icon-to-label gap |
| `p-3` / `gap-3` | 12px | List item padding |
| `p-4` / `gap-4` | 16px | Card internal padding |
| `p-6` / `gap-6` | 24px | Section padding |
| `p-8` / `gap-8` | 32px | Section gaps |
| `p-12` | 48px | Large section dividers |
| `p-16` | 64px | Page top/bottom padding |

### Border Radius

| Class | Value | Usage |
|---|---|---|
| `rounded-xs` | 2px | Tiny decorative elements |
| `rounded-sm` | 4px | Checkboxes, tags |
| `rounded-md` | 8px | Buttons, inputs, badges |
| `rounded-lg` | 12px | Cards, panels |
| `rounded-xl` | 16px | Large containers |
| `rounded-2xl` | 24px | Oversized elements |
| `rounded-full` | pill | Badge chips, avatars |

### Shadows

All shadows use warm-tinted RGBA values:

| Class | Usage |
|---|---|
| `shadow-xs` | Subtle card elevation (default card state) |
| `shadow-sm` | Hovered cards |
| `shadow-md` | Dropdowns, tooltips |
| `shadow-lg` | Modals, floating panels |
| `shadow-focus` | Focus ring (used via `focus-ring` utility) |

---

## 5. Custom Utilities

Defined in `frontend/src/app.css` via `@utility`:

| Class | Effect |
|---|---|
| `mode-value` | Sets `--mode-accent` to green-700, `--mode-accent-light` to green-100, `--mode-accent-bg` to green-50 |
| `mode-dividend` | Sets `--mode-accent` to gold-500, `--mode-accent-light` to gold-100, `--mode-accent-bg` to gold-50 |
| `text-profit` | `color: var(--fin-profit)` — adapts to theme |
| `text-loss` | `color: var(--fin-loss)` — adapts to theme |
| `focus-ring` | Adds `box-shadow: var(--shadow-focus)` on `:focus-visible`, clears outline |
| `skeleton` | Animated shimmer loading placeholder (1.5s infinite) — use via `SkeletonLine`, `SkeletonCard`, `SkeletonTable` components |
| `transition-fast` | 120ms `cubic-bezier(0.4, 0, 0.2, 1)` transition |
| `transition-normal` | 200ms `cubic-bezier(0.4, 0, 0.2, 1)` transition |
| `transition-slow` | 320ms `cubic-bezier(0.4, 0, 0.2, 1)` transition |

---

## 6. Component Patterns

Import paths use relative paths from the importing file's location. There is no SvelteKit path alias (`$lib`) in this Wails app — use relative imports.

### Button

```svelte
<script lang="ts">
  import Button from "../components/Button.svelte";
</script>

<Button variant="primary">Buy 3 lots</Button>
<Button variant="secondary">Cancel</Button>
<Button variant="ghost">Skip</Button>
<Button variant="danger" size="sm">Sell</Button>
<Button variant="gold">Add to Dividend</Button>
<Button variant="primary" loading>Saving...</Button>
```

Props:
- `variant`: `"primary"` | `"secondary"` | `"ghost"` | `"danger"` | `"gold"` (default: `"primary"`)
- `size`: `"sm"` | `"md"` | `"lg"` (default: `"md"`)
- `disabled`: `boolean` (default: `false`)
- `loading`: `boolean` — shows spinner, disables interaction (default: `false`)
- `type`: `"button"` | `"submit"` | `"reset"` (default: `"button"`)
- `onclick`: `(e: MouseEvent) => void`
- `children`: Snippet (required)

Variant styles: `primary` = green filled, `secondary` = outlined, `ghost` = transparent, `danger` = red filled, `gold` = gold filled.

### Badge

```svelte
<script lang="ts">
  import Badge from "../components/Badge.svelte";
</script>

<Badge variant="value">Value Mode</Badge>
<Badge variant="dividend">Dividend</Badge>
<Badge variant="profit">+5.82%</Badge>
<Badge variant="loss">-2.15%</Badge>
<Badge variant="warning">Below Fair Value</Badge>
```

Props:
- `variant`: `"value"` | `"dividend"` | `"profit"` | `"loss"` | `"warning"` (default: `"value"`)
- `children`: Snippet (required)

Always a pill shape (`rounded-full`). Renders as `<span>` — inline in text flow.

### Alert

```svelte
<script lang="ts">
  import Alert from "../components/Alert.svelte";
</script>

<Alert variant="positive">BBRI has reached your entry target.</Alert>
<Alert variant="warning" dismissible>Payout ratio exceeded 85%.</Alert>
<Alert variant="negative">Fetch failed — check your connection.</Alert>
<Alert variant="info">Data last updated 5 minutes ago.</Alert>
```

Props:
- `variant`: `"positive"` | `"warning"` | `"negative"` | `"info"` (default: `"info"`)
- `dismissible`: `boolean` — shows an X button that hides the alert (default: `false`)
- `children`: Snippet (required)

Dismissal is local state — the component hides itself. No callback is emitted.

### ThemeToggle

```svelte
<script lang="ts">
  import ThemeToggle from "../components/ThemeToggle.svelte";
</script>

<ThemeToggle />
<!-- Cycles: light → dark → system. Persists preference to localStorage. -->
```

No props. Reads and writes the `theme` store directly. Shows a sun icon (light), moon icon (dark), or monitor icon (system).

Import the store to read theme state elsewhere:

```ts
import { theme } from "../stores/theme.svelte";

// Read
theme.current;      // "light" | "dark"
theme.preference;   // "light" | "dark" | "system"
theme.isDark;       // boolean

// Write
theme.set("dark");
theme.toggle();     // cycles light → dark → system
```

### ModeTabs

```svelte
<script lang="ts">
  import ModeTabs from "../components/ModeTabs.svelte";
</script>

<ModeTabs />
<!-- Auto-connects to mode store. Switches accent color globally. -->
```

No props. Reads and writes the `mode` store directly. Renders as a `role="tablist"` with two tab buttons.

Import the store to read mode state elsewhere:

```ts
import { mode } from "../stores/mode.svelte";

// Read
mode.current;         // "value" | "dividend"
mode.config;          // { label, emoji, accent, accentLight, badgeClass, containerClass }
mode.isValue;         // boolean
mode.isDividend;      // boolean
mode.accentColor;     // CSS color string, e.g. "var(--color-green-700)"
mode.containerClass;  // "mode-value" or "mode-dividend"
mode.badgeClass;      // Tailwind classes for the active mode badge

// Write
mode.set("dividend");
mode.toggle();        // switches between value and dividend
```

### StockCard

```svelte
<script lang="ts">
  import StockCard from "../components/StockCard.svelte";
</script>

<StockCard
  ticker="BBRI"
  name="Bank Rakyat Indonesia"
  price={4250}
  change={125}
  changePercent={3.03}
  mode="value"
  metrics={[
    { label: "PBV", value: "1.28×" },
    { label: "Graham #", value: "Rp 5,180" },
    { label: "Upside", value: "+21.9%", positive: true },
  ]}
/>
```

Props:
- `ticker`: `string` (required) — uppercase stock code
- `name`: `string` (required) — full company name
- `price`: `number` (required) — current price in IDR
- `change`: `number` (required) — absolute price change (negative = loss)
- `changePercent`: `number` (required) — percentage change
- `mode`: `"value"` | `"dividend"` (default: `"value"`)
- `metrics`: `{ label: string; value: string; positive?: boolean }[]` (default: `[]`)

The `metrics` array renders in a 3-column grid. Set `positive: true` for green text, `positive: false` for red text, omit for default text color.

Price and change use `price.toLocaleString("id-ID")` and `changePercent.toFixed(2)` for Indonesian number formatting.

---

## 7. Layout Structure

```
+--------------------------------------------------+
|  Sidebar (220px)          |  Main Content (fluid) |
|  bg-bg-secondary          |  bg-bg-primary        |
|                           |                       |
|  panen (logo)             |  +----------------+   |
|                           |  |  Summary Cards  |   |
|  [ModeTabs]               |  |  bg-bg-elevated |   |
|                           |  +----------------+   |
|  Overview  <- active      |                       |
|  Screener                 |  +----------------+   |
|  Watchlist                |  |  Data Table     |   |
|  Portfolio                |  |  bg-bg-elevated |   |
|  Monthly Plan             |  +----------------+   |
|                           |                       |
|  ---                      |                       |
|  Settings                 |                       |
|  [ThemeToggle]            |                       |
+--------------------------------------------------+
```

Three-layer depth: `bg-secondary` (sidebar) → `bg-primary` (canvas) → `bg-elevated` (cards).

The sidebar width is `w-sidebar` (220px) — defined as `--width-sidebar` in `@theme`.

Sidebar nav item pattern:

```svelte
<a
  href="/overview"
  class="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium
    text-text-secondary hover:bg-bg-tertiary hover:text-text-primary transition-fast
    {isActive ? 'bg-green-100 text-green-800 font-semibold' : ''}"
>
  <span aria-hidden="true">icon</span> Overview
</a>
```

---

## 8. Motion Rules

| Context | Duration | Utility |
|---|---|---|
| Hover, focus, toggle | 120ms | `transition-fast` |
| Dropdown, tooltip, card expand | 200ms | `transition-normal` |
| Mode switch, modal, page transition | 320ms | `transition-slow` |
| Profit/loss color change | instant | No transition class |
| Skeleton loading | 1.5s infinite | `skeleton` utility |

Never animate numbers counting up — it undermines trust in a financial context.

The transition utilities set `transition-duration` and `transition-timing-function`. Add `transition-property` (e.g., `transition-colors`, `transition-all`) as needed:

```svelte
<button class="transition-colors transition-fast hover:bg-bg-tertiary">
  Click me
</button>
```

---

## 9. Iconography

**Library:** [Lucide](https://lucide.dev) via `lucide-svelte` — tree-shakeable Svelte components.

### Usage

```svelte
<script lang="ts">
  import { Search, Settings, LoaderCircle } from "lucide-svelte";
</script>

<!-- Navigation / UI icon -->
<Search size={20} strokeWidth={1.5} />

<!-- Spinner -->
<LoaderCircle size={16} strokeWidth={2} class="animate-spin" />
```

### Sizing Conventions

| Context | Size | Stroke Width |
|---|---|---|
| Navigation icons (sidebar) | 20 | 1.5 |
| Theme toggle, UI actions | 20 | 1.5 |
| Close / dismiss buttons | 16 | 2 |
| Loading spinners (inline) | 16 | 2 |
| Loading spinners (page-level) | 20 | 2 |

### Guidelines

- Import only the icons you need — each is a separate module for tree-shaking
- Use `class` prop to forward Tailwind classes (e.g., `class="animate-spin"`, `class="shrink-0"`)
- **Emoji** remains acceptable for modes (`ModeTabs.svelte`) and decorative content (`StockCard.svelte`)
- Always pair icons with a visible text label
- Never communicate state through color alone — combine with icon and text

---

## 10. Do / Don't

### Do

- Use `font-mono` for all financial numbers
- Right-align numbers in tables
- Show suggestions as actions: "Buy 3 lots at Rp 4,200"
- Use warm off-white backgrounds (`bg-primary: #fefcf7`), not pure white
- Pair emoji icons with text labels (emoji `aria-hidden="true"`)
- Use green for Value Mode, gold for Dividend Mode — keep them distinct
- Apply profit/loss color changes instantly (no transition)
- Use the `skeleton` utility for loading states
- Add last-updated timestamps on all data displays
- Use relative import paths: `import { theme } from "../stores/theme.svelte"`

### Don't

- Use real-time tickers or flashy price animations
- Mix brand green with unrelated blues or purples
- Show raw data without context or a suggested action
- Use spinners for content loading (use skeletons instead)
- Rely on color alone for state — always add icon or text
- Use pure white (`#ffffff`) or pure black (`#000000`) as page backgrounds
- Add gradients to buttons or cards
- Display financial data without a timestamp
- Use SvelteKit-specific APIs (`$app/environment`, `$lib` alias) — this is a Wails app
- Use `sessionStorage` for application state — use Go backend for persistence; `localStorage` is acceptable only for UI preferences (theme)
