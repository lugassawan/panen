<script lang="ts">
import LoadingState from "../LoadingState.svelte";

let {
  message,
  size,
  className,
}: {
  message?: string;
  size?: "sm" | "md";
  className?: string;
} = $props();
</script>

<LoadingState {message} {size} class={className} />
