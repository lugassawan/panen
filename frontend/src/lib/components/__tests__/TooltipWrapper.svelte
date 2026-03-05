<script lang="ts">
import type { Snippet } from "svelte";
import Tooltip from "../Tooltip.svelte";

let {
  text = "Tooltip text",
  position,
  children,
}: {
  text?: string;
  position?: "top" | "bottom" | "left" | "right";
  children?: Snippet;
} = $props();
</script>

<Tooltip {text} {position}>
  <button>Hover me</button>
</Tooltip>
