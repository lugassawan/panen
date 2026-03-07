<script lang="ts">
import SearchableSelect from "./SearchableSelect.svelte";

type Item = { id: string; label: string };

let {
  items = [],
  value = $bindable(""),
  placeholder = "",
  onselect,
  showFooter = false,
}: {
  items?: Item[];
  value?: string;
  placeholder?: string;
  onselect?: (key: string) => void;
  showFooter?: boolean;
} = $props();

function filterFn(item: Item, query: string): boolean {
  return item.label.toLowerCase().includes(query.toLowerCase());
}

function displayFn(item: Item): string {
  return item.label;
}

function keyFn(item: Item): string {
  return item.id;
}
</script>

{#if showFooter}
  <SearchableSelect
    {items}
    bind:value
    {filterFn}
    {displayFn}
    {keyFn}
    {placeholder}
    {onselect}
  >
    {#snippet children({ item })}
      <span>{item.label}</span>
    {/snippet}
    {#snippet footer()}
      <button type="button">Footer action</button>
    {/snippet}
  </SearchableSelect>
{:else}
  <SearchableSelect
    {items}
    bind:value
    {filterFn}
    {displayFn}
    {keyFn}
    {placeholder}
    {onselect}
  >
    {#snippet children({ item })}
      <span>{item.label}</span>
    {/snippet}
  </SearchableSelect>
{/if}
