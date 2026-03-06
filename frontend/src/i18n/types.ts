export type Locale = "en" | "id";

export interface Translations {
  [key: string]: string | Translations;
}
