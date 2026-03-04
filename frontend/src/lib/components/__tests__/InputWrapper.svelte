<script lang="ts">
import Input from "../Input.svelte";

let {
  id,
  type,
  placeholder,
  value = $bindable(""),
  disabled,
  min,
  max,
  step,
  "aria-label": ariaLabel,
  oninput,
  onchange,
  class: extraClass,
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

<Input {id} {type} {placeholder} bind:value {disabled} {min} {max} {step} aria-label={ariaLabel} {oninput} {onchange} class={extraClass} />
