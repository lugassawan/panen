<script lang="ts">
import { formatRelativeTime } from "../format";

let {
  date,
  label = "Last updated",
  class: className = "",
}: {
  date: string | Date;
  label?: string;
  class?: string;
} = $props();

let isoString = $derived(typeof date === "string" ? date : date.toISOString());
let relative = $state("");

$effect(() => {
  relative = formatRelativeTime(isoString);
  const interval = setInterval(() => {
    relative = formatRelativeTime(isoString);
  }, 60000);
  return () => clearInterval(interval);
});
</script>

<span class="text-xs text-text-muted {className}">
  {label}: <time datetime={isoString}>{relative}</time>
</span>
