<script lang="ts">
let {
  id,
  type = "text",
  placeholder,
  value = $bindable<string | number>(),
  disabled = false,
  min,
  max,
  step,
  "aria-label": ariaLabel,
  oninput,
  onchange,
  class: extraClass = "",
}: {
  id?: string;
  type?: string;
  placeholder?: string;
  value?: string | number;
  disabled?: boolean;
  min?: string | number;
  max?: string | number;
  step?: string | number;
  "aria-label"?: string;
  oninput?: (e: Event & { currentTarget: HTMLInputElement }) => void;
  onchange?: (e: Event & { currentTarget: HTMLInputElement }) => void;
  class?: string;
} = $props();
</script>

<input
  {id}
  {type}
  {placeholder}
  bind:value
  {disabled}
  {min}
  {max}
  {step}
  aria-label={ariaLabel}
  {oninput}
  {onchange}
  class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary outline-none focus:border-green-700 focus-ring {extraClass}"
/>
