<script lang="ts">
import { toastStore } from "../../stores/toast.svelte";
import ToastContainer from "../ToastContainer.svelte";

let {
  toasts = [],
}: {
  toasts?: Array<{ message: string; variant: "success" | "error" | "warning" | "info" }>;
} = $props();

for (const t of toasts) {
  toastStore.add(t.message, t.variant);
}
</script>

<ToastContainer />
