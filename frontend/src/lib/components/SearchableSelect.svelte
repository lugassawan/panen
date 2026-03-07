<script lang="ts" generics="T">
import { ChevronDown, Search } from "lucide-svelte";
import type { Snippet } from "svelte";
import { t } from "../../i18n";

let {
  items,
  value = $bindable(""),
  filterFn,
  displayFn,
  keyFn,
  placeholder = "",
  disabled = false,
  id,
  onselect,
  fallbackDisplay = "",
  children,
  footer,
}: {
  items: T[];
  value?: string;
  filterFn: (item: T, query: string) => boolean;
  displayFn: (item: T) => string;
  keyFn: (item: T) => string;
  placeholder?: string;
  disabled?: boolean;
  id?: string;
  onselect?: (key: string) => void;
  fallbackDisplay?: string;
  children: Snippet<[{ item: T; active: boolean }]>;
  footer?: Snippet;
} = $props();

let open = $state(false);
let query = $state("");
let activeIndex = $state(0);
let containerEl = $state<HTMLElement | null>(null);
let inputEl = $state<HTMLInputElement | null>(null);
let listEl = $state<HTMLElement | null>(null);

let filtered = $derived(query ? items.filter((item) => filterFn(item, query)) : items);

let selectedItem = $derived(items.find((item) => keyFn(item) === value));
let displayText = $derived(selectedItem ? displayFn(selectedItem) : fallbackDisplay);

function openDropdown() {
  if (disabled) return;
  open = true;
  query = "";
  activeIndex = 0;
}

function closeDropdown() {
  open = false;
  query = "";
}

function selectItem(key: string) {
  value = key;
  closeDropdown();
  onselect?.(key);
}

function handleKeydown(e: KeyboardEvent) {
  if (!open) {
    if (e.key === "ArrowDown" || e.key === "Enter") {
      e.preventDefault();
      openDropdown();
    }
    return;
  }

  const totalItems = filtered.length + (footer ? 1 : 0);

  switch (e.key) {
    case "ArrowDown":
      e.preventDefault();
      activeIndex = (activeIndex + 1) % totalItems;
      scrollToActive();
      break;
    case "ArrowUp":
      e.preventDefault();
      activeIndex = (activeIndex - 1 + totalItems) % totalItems;
      scrollToActive();
      break;
    case "Enter":
      e.preventDefault();
      if (activeIndex < filtered.length) {
        selectItem(keyFn(filtered[activeIndex]));
      }
      break;
    case "Escape":
      e.preventDefault();
      closeDropdown();
      break;
  }
}

function scrollToActive() {
  requestAnimationFrame(() => {
    const active = listEl?.querySelector('[data-active="true"]');
    active?.scrollIntoView({ block: "nearest" });
  });
}

function handleDocumentClick(e: MouseEvent) {
  if (containerEl && !containerEl.contains(e.target as Node)) {
    closeDropdown();
  }
}

$effect(() => {
  if (open) {
    document.addEventListener("click", handleDocumentClick);
    return () => document.removeEventListener("click", handleDocumentClick);
  }
});
</script>

<div bind:this={containerEl} class="relative">
  <div class="relative">
    <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-text-tertiary">
      <Search size={14} strokeWidth={2} />
    </div>
    <input
      {id}
      type="text"
      role="combobox"
      aria-expanded={open}
      aria-haspopup="listbox"
      aria-autocomplete="list"
      {disabled}
      placeholder={open ? placeholder : displayText || placeholder}
      value={open ? query : displayText}
      onfocus={openDropdown}
      oninput={(e) => {
        query = e.currentTarget.value;
        activeIndex = 0;
        if (!open) openDropdown();
      }}
      onkeydown={handleKeydown}
      bind:this={inputEl}
      class="w-full rounded border border-border-default bg-bg-elevated py-2 pr-8 pl-8 text-sm text-text-primary outline-none transition-fast hover:border-border-strong focus-border disabled:opacity-60 focus-ring {!open && !value ? 'text-text-muted' : ''}"
    />
    <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2 text-text-tertiary">
      <ChevronDown size={16} strokeWidth={2} class={open ? "rotate-180 transition-fast" : "transition-fast"} />
    </div>
  </div>

  {#if open}
    <ul
      role="listbox"
      bind:this={listEl}
      class="absolute z-10 mt-1 max-h-56 w-full overflow-y-auto rounded border border-border-default bg-bg-elevated shadow-lg"
    >
      {#if filtered.length === 0}
        <li class="px-3 py-2 text-sm text-text-muted">{t("common.noResults")}</li>
      {:else}
        {#each filtered as item, i}
          <li
            role="option"
            aria-selected={keyFn(item) === value}
            data-active={i === activeIndex}
            onmouseenter={() => (activeIndex = i)}
            onclick={() => selectItem(keyFn(item))}
            class="cursor-pointer px-3 py-2 text-sm {i === activeIndex ? 'bg-bg-tertiary' : ''}"
          >
            {@render children({ item, active: i === activeIndex })}
          </li>
        {/each}
      {/if}
      {#if footer}
        <li class="border-t border-border-default">
          {@render footer()}
        </li>
      {/if}
    </ul>
  {/if}
</div>
