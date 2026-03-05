<script lang="ts">
import BrokerageAccountForm from "../../components/BrokerageAccountForm.svelte";
import type { BrokerageAccountResponse, BrokerConfigResponse } from "../../lib/types";
import PortfolioForm from "./PortfolioForm.svelte";

interface Props {
  brokerConfigs: BrokerConfigResponse[];
  brokerageAcctId: string | null;
  showPortfolioForm: boolean;
  onBrokerageCreated: (acct: BrokerageAccountResponse) => void;
  onPortfolioCreated: () => void;
}

let {
  brokerConfigs,
  brokerageAcctId,
  showPortfolioForm,
  onBrokerageCreated,
  onPortfolioCreated,
}: Props = $props();
</script>

<div class="mx-auto max-w-lg">
  {#if !showPortfolioForm}
    <h2 class="mb-6 text-xl font-semibold text-text-primary">Set Up Your Brokerage</h2>
    <div class="rounded border border-border-default bg-bg-elevated p-6">
      <BrokerageAccountForm
        {brokerConfigs}
        onSaved={onBrokerageCreated}
      />
    </div>
  {:else}
    <h2 class="mb-6 text-xl font-semibold text-text-primary">Create Your Portfolio</h2>
    <div class="rounded border border-border-default bg-bg-elevated p-6">
      <PortfolioForm
        brokerageAcctId={brokerageAcctId ?? ""}
        onSaved={onPortfolioCreated}
      />
    </div>
  {/if}
</div>
