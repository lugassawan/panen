<script lang="ts">
let {
  width = "100%",
  height = "1rem",
}: {
  width?: string;
  height?: string;
} = $props();
</script>

<div class="skeleton" style="width: {width}; height: {height}"></div>
