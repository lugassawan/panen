<script lang="ts">
import Button from "../../lib/components/Button.svelte";
import Select from "../../lib/components/Select.svelte";

let { onSave }: { onSave: (day: number) => void } = $props();

let selectedDay = $state<number>(25);

function handleSave() {
  onSave(selectedDay);
}
</script>

<div class="flex flex-col items-center justify-center gap-6 py-16">
  <div class="w-full max-w-sm rounded-lg border border-border-default bg-bg-elevated p-6">
    <h2 class="text-lg font-semibold text-text-primary font-display">Set Your Payday</h2>
    <p class="mt-2 text-sm text-text-secondary">
      Choose the day of the month when you receive your salary. This helps track your monthly investment schedule.
    </p>

    <div class="mt-6">
      <label for="payday-day" class="mb-1.5 block text-sm font-medium text-text-secondary">
        Payday Day
      </label>
      <Select id="payday-day" bind:value={selectedDay} aria-label="Select payday day">
        {#each Array.from({ length: 31 }, (_, i) => i + 1) as day}
          <option value={day}>{day}</option>
        {/each}
      </Select>
    </div>

    <div class="mt-6">
      <Button variant="primary" onclick={handleSave}>Save Payday</Button>
    </div>
  </div>
</div>
