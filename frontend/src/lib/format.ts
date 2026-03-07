import { locale, t } from "../i18n";

function intlLocale(): string {
  return locale.current === "id" ? "id-ID" : "en-US";
}

const cache = new Map<string, Intl.NumberFormat>();

function getFormatter(options: Intl.NumberFormatOptions): Intl.NumberFormat {
  const key = `${intlLocale()}:${JSON.stringify(options)}`;
  let fmt = cache.get(key);
  if (!fmt) {
    fmt = new Intl.NumberFormat(intlLocale(), options);
    cache.set(key, fmt);
  }
  return fmt;
}

export function formatRupiah(value: number): string {
  return getFormatter({
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value);
}

export function formatDecimal(value: number, digits = 2): string {
  return getFormatter({
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value);
}

export function formatPercent(value: number, digits = 2): string {
  return `${formatDecimal(value, digits)}%`;
}

export function formatRelativeTime(isoString: string): string {
  if (!isoString) return t("format.notSynced");
  const date = new Date(isoString);
  const now = Date.now();
  const diffMs = now - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return t("format.justNow");
  if (diffMin < 60) return t("format.minutesAgo", { count: diffMin });
  const diffHrs = Math.floor(diffMin / 60);
  if (diffHrs < 24) return t("format.hoursAgo", { count: diffHrs });
  const diffDays = Math.floor(diffHrs / 24);
  return t("format.daysAgo", { count: diffDays });
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  const units = ["KB", "MB", "GB"];
  let value = bytes;
  let unitIndex = -1;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex++;
  }
  return `${value.toFixed(1)} ${units[unitIndex]}`;
}

export function formatDate(isoString: string): string {
  if (!isoString) return "";
  const date = new Date(isoString);
  return new Intl.DateTimeFormat(intlLocale(), {
    year: "numeric",
    month: "long",
    day: "numeric",
  }).format(date);
}
