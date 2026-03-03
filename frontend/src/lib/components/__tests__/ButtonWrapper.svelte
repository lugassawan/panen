<script lang="ts">
import Button from "../Button.svelte";

let {
  text = "Click me",
  variant,
  size,
  disabled,
  loading,
  type,
  onclick,
}: {
  text?: string;
  variant?: "primary" | "secondary" | "ghost" | "danger" | "gold";
  size?: "sm" | "md" | "lg";
  disabled?: boolean;
  loading?: boolean;
  type?: "button" | "submit" | "reset";
  onclick?: (e: MouseEvent) => void;
} = $props();
</script>

<Button {variant} {size} {disabled} {loading} {type} {onclick}>
  {text}
</Button>
