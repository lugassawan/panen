<script lang="ts">
import StockCard from "../StockCard.svelte";

let {
  ticker = "BBCA",
  name = "Bank Central Asia",
  price = 9500,
  change = 100,
  changePercent = 1.06,
  mode,
  metrics,
}: {
  ticker?: string;
  name?: string;
  price?: number;
  change?: number;
  changePercent?: number;
  mode?: "value" | "dividend";
  metrics?: { label: string; value: string; positive?: boolean }[];
} = $props();
</script>

<StockCard {ticker} {name} {price} {change} {changePercent} {mode} {metrics} />
