<script lang="ts">
import Alert from "../Alert.svelte";

let {
  text = "Alert message",
  variant,
  dismissible,
}: {
  text?: string;
  variant?: "positive" | "warning" | "negative" | "info";
  dismissible?: boolean;
} = $props();
</script>

<Alert {variant} {dismissible}>
  {text}
</Alert>
