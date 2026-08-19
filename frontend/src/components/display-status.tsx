import type { DisplayState } from "../types/display";
import { useTranslation } from "../hooks/use-translation";

type DisplayStatusProps = {
  display: DisplayState;
};

export function DisplayStatus({ display }: DisplayStatusProps) {
  const { t } = useTranslation();

  return (
    <section>
      <h2>{t("display.title")}</h2>

      <p>{t("display.power", { power: display.power })}</p>

      <p>{t("display.brightness", { brightness: display.brightness })}</p>
    </section>
  );
}
