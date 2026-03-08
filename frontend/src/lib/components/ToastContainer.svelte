<script lang="ts">
import { toastStore } from "../stores/toast.svelte";
import Toast from "./Toast.svelte";
</script>

{#if toastStore.toasts.length > 0}
  <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 w-80" role="status" aria-live="polite">
    {#each toastStore.toasts as toast (toast.id)}
      <Toast
        message={toast.message}
        variant={toast.variant}
        onDismiss={() => toastStore.dismiss(toast.id)}
      />
    {/each}
  </div>
{/if}
