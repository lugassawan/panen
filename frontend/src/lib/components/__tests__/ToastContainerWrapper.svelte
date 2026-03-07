<script lang="ts">
import { untrack } from "svelte";
import { toastStore } from "../../stores/toast.svelte";
import ToastContainer from "../ToastContainer.svelte";

let {
  toasts = [],
}: {
  toasts?: Array<{ message: string; variant: "success" | "error" | "warning" | "info" }>;
} = $props();

$effect(() => {
  const items = toasts;
  untrack(() => {
    for (const t of items) {
      toastStore.add(t.message, t.variant);
    }
  });
});
</script>

<ToastContainer />
