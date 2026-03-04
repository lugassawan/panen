<script lang="ts">
import Select from "../Select.svelte";

let {
  id,
  value = $bindable(""),
  disabled,
  "aria-label": ariaLabel,
  onchange,
  class: extraClass,
}: {
  id?: string;
  value?: string | number;
  disabled?: boolean;
  "aria-label"?: string;
  onchange?: (e: Event & { currentTarget: HTMLSelectElement }) => void;
  class?: string;
} = $props();
</script>

<Select {id} bind:value {disabled} aria-label={ariaLabel} {onchange} class={extraClass}>
  <option value="a">Option A</option>
  <option value="b">Option B</option>
  <option value="c">Option C</option>
</Select>
