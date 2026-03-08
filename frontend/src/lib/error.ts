import { t } from "../i18n";

/**
 * Formats a backend error string for display.
 * Backend AppErrors have format "ERR_CODE|message".
 * If a translation exists for the code, use it; otherwise fall back to raw message.
 */
export function formatError(raw: string): string {
  const idx = raw.indexOf("|");
  if (idx > 0) {
    const code = raw.substring(0, idx);
    if (code.startsWith("ERR_")) {
      const key = `error.${code}`;
      const translated = t(key);
      if (translated !== key) return translated;
    }
  }
  return raw;
}
