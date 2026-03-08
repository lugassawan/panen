<script lang="ts">
import { untrack } from "svelte";
import { t } from "../../i18n";
import Badge from "../../lib/components/Badge.svelte";
import Button from "../../lib/components/Button.svelte";
import Modal from "../../lib/components/Modal.svelte";
import type { DiagnosticResponse, DiagnosticSignal } from "../../lib/types";

let {
  ticker,
  diagnostic,
  onUpdate,
  onClose,
}: {
  ticker: string;
  diagnostic: DiagnosticResponse;
  onUpdate: (companyBadNews: boolean | null, fundamentalsOK: boolean | null) => void;
  onClose: () => void;
} = $props();

let companyBadNews = $state<boolean | null>(untrack(() => diagnostic.companyBadNews));
let fundamentalsOK = $state<boolean | null>(untrack(() => diagnostic.fundamentalsOK));

function toggleCompanyBadNews() {
  companyBadNews = companyBadNews === null ? true : companyBadNews ? false : null;
  onUpdate(companyBadNews, fundamentalsOK);
}

function toggleFundamentalsOK() {
  fundamentalsOK = fundamentalsOK === null ? true : fundamentalsOK ? false : null;
  onUpdate(companyBadNews, fundamentalsOK);
}

const signalConfig: Record<
  DiagnosticSignal,
  { variant: "profit" | "loss" | "warning"; label: string }
> = {
  OPPORTUNITY: { variant: "profit", label: t("crashPlaybook.opportunity") },
  FALLING_KNIFE: { variant: "loss", label: t("crashPlaybook.fallingKnife") },
  INCONCLUSIVE: { variant: "warning", label: t("crashPlaybook.inconclusive") },
};

const signal = $derived(signalConfig[diagnostic.signal]);

function checkColor(value: boolean | null, invertMeaning = false): string {
  if (value === null) return "text-text-secondary";
  const pass = invertMeaning ? !value : value;
  return pass ? "text-profit" : "text-loss";
}
</script>

<Modal aria-label={t("crashPlaybook.fallingKnifeTitle")} onClose={onClose} size="md">
  <div class="flex items-center justify-between">
    <h3 class="font-display text-lg font-semibold text-text-primary">{t("crashPlaybook.fallingKnifeTitle")}</h3>
    <Badge variant={signal.variant}>{signal.label}</Badge>
  </div>
  <p class="mt-1 text-sm text-text-secondary">
    {t("crashPlaybook.evaluatingTicker", { ticker })}
  </p>

  <div class="mt-4 space-y-3">
    <div class="flex items-center justify-between rounded-md bg-bg-primary px-3 py-2">
      <span class="text-sm text-text-secondary">{t("crashPlaybook.marketCrashed")}</span>
      <span class="font-mono text-sm font-medium {diagnostic.marketCrashed ? 'text-loss' : 'text-profit'}">
        {diagnostic.marketCrashed ? t("common.yes") : t("common.no")}
      </span>
    </div>

    <button
      class="flex w-full items-center justify-between rounded-md bg-bg-primary px-3 py-2 transition-fast hover:bg-bg-tertiary focus-ring"
      onclick={toggleCompanyBadNews}
    >
      <span class="text-sm text-text-secondary">{t("crashPlaybook.companyNews")}</span>
      <span class="font-mono text-sm font-medium {checkColor(companyBadNews, true)}">
        {companyBadNews === null ? t("common.unknown") : companyBadNews ? t("common.yes") : t("common.no")}
      </span>
    </button>

    <button
      class="flex w-full items-center justify-between rounded-md bg-bg-primary px-3 py-2 transition-fast hover:bg-bg-tertiary focus-ring"
      onclick={toggleFundamentalsOK}
    >
      <span class="text-sm text-text-secondary">{t("crashPlaybook.fundamentalsHealthy")}</span>
      <span class="font-mono text-sm font-medium {checkColor(fundamentalsOK)}">
        {fundamentalsOK === null ? t("common.unknown") : fundamentalsOK ? t("common.yes") : t("common.no")}
      </span>
    </button>

    <div class="flex items-center justify-between rounded-md bg-bg-primary px-3 py-2">
      <span class="text-sm text-text-secondary">{t("crashPlaybook.belowEntry")}</span>
      <span class="font-mono text-sm font-medium {diagnostic.belowEntry ? 'text-profit' : 'text-text-secondary'}">
        {diagnostic.belowEntry ? t("common.yes") : t("common.no")}
      </span>
    </div>
  </div>

  <div class="mt-4 flex justify-end">
    <Button variant="secondary" size="sm" onclick={onClose}>{t("common.close")}</Button>
  </div>
</Modal>
