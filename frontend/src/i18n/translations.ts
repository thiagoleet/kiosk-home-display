import enUS from "./locales/en-US.json";
import ptBR from "./locales/pt-BR.json";

export type Locale = "pt-BR" | "en-US";
export type TranslationKey = keyof typeof ptBR;
export type TranslationValues = Record<string, string | number>;

const translations = {
  "pt-BR": ptBR,
  "en-US": enUS,
} satisfies Record<Locale, Record<TranslationKey, string>>;

export function translate(
  locale: Locale,
  key: TranslationKey,
  values: TranslationValues = {},
) {
  return translations[locale][key].replace(
    /{{(\w+)}}/g,
    (_, name: string) => String(values[name] ?? ""),
  );
}
