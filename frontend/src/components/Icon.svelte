<script lang="ts">
let {
  path,
  size = "h-5 w-5",
}: {
  path: string;
  size?: string;
} = $props();
</script>

<svg
  class="{size} shrink-0"
  fill="none"
  viewBox="0 0 24 24"
  stroke-width="1.5"
  stroke="currentColor"
>
  <path stroke-linecap="round" stroke-linejoin="round" d={path} />
</svg>
