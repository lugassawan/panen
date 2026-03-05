<script lang="ts">
import type { Snippet } from "svelte";

let {
  text,
  position = "top",
  children,
}: {
  text: string;
  position?: "top" | "bottom" | "left" | "right";
  children: Snippet;
} = $props();

let visible = $state(false);
const tooltipId = `tooltip-${Math.random().toString(36).slice(2, 9)}`;

function show() {
  visible = true;
}

function hide() {
  visible = false;
}

const positionClasses: Record<string, string> = {
  top: "bottom-full left-1/2 -translate-x-1/2 mb-2",
  bottom: "top-full left-1/2 -translate-x-1/2 mt-2",
  left: "right-full top-1/2 -translate-y-1/2 mr-2",
  right: "left-full top-1/2 -translate-y-1/2 ml-2",
};
</script>

<div
  class="relative inline-flex"
  role="presentation"
  onmouseenter={show}
  onmouseleave={hide}
  onfocusin={show}
  onfocusout={hide}
>
  <div aria-describedby={visible ? tooltipId : undefined}>
    {@render children()}
  </div>

  {#if visible}
    <div
      id={tooltipId}
      role="tooltip"
      class="absolute z-50 whitespace-nowrap rounded-md bg-bg-elevated px-2.5 py-1.5 text-xs text-text-primary shadow-md transition-fast pointer-events-none {positionClasses[position]}"
    >
      {text}
    </div>
  {/if}
</div>
