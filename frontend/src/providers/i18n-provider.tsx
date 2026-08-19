import { useCallback, useMemo, type PropsWithChildren } from "react";

import { useKiosk } from "../hooks/use-kiosk";
import { translate, type TranslationKey, type TranslationValues } from "../i18n/translations";
import { I18nContext, type I18nContextValue } from "./i18n-context";

export function I18nProvider({ children }: PropsWithChildren) {
  const { profile } = useKiosk();
  const { locale } = profile;

  const t = useCallback(
    (key: TranslationKey, values?: TranslationValues) =>
      translate(locale, key, values),
    [locale],
  );

  const value = useMemo<I18nContextValue>(
    () => ({ locale, t }),
    [locale, t],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}
