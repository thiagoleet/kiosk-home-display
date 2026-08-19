import { createContext } from "react";

import type {
  Locale,
  TranslationKey,
  TranslationValues,
} from "../i18n/translations";

export type I18nContextValue = {
  locale: Locale;
  t: (key: TranslationKey, values?: TranslationValues) => string;
};

export const I18nContext = createContext<I18nContextValue | null>(null);
