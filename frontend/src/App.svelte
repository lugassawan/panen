<script lang="ts">
import { Greet } from "../wailsjs/go/app/App";

let name = $state("");
let greeting = $state("");

async function greet() {
  greeting = await Greet(name);
}
</script>

<main class="flex min-h-screen items-center justify-center bg-neutral-950 text-white">
  <div class="flex flex-col items-center gap-6">
    <h1 class="text-4xl font-bold">Panen</h1>
    <p class="text-neutral-400">Desktop decision engine for IDX investors</p>
    <div class="flex gap-2">
      <input
        bind:value={name}
        placeholder="Enter your name"
        class="rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm outline-none focus:border-amber-500"
      />
      <button
        onclick={greet}
        class="rounded bg-amber-600 px-4 py-2 text-sm font-medium hover:bg-amber-500"
      >
        Greet
      </button>
    </div>
    {#if greeting}
      <p class="text-amber-400">{greeting}</p>
    {/if}
  </div>
</main>
