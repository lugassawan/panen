<script lang="ts">
import Badge from "../Badge.svelte";

let {
  text = "Badge",
  variant,
}: {
  text?: string;
  variant?: "value" | "dividend" | "profit" | "loss" | "warning";
} = $props();
</script>

<Badge {variant}>
  {text}
</Badge>
