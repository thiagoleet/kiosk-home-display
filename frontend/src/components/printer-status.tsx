import type { PrintJob } from "../types/printer";
import type { PrinterState } from "../hooks/use-printer";
import { useTranslation } from "../hooks/use-translation";

type PrinterStatusProps = {
  state: PrinterState;
  job: PrintJob | null;
};

export function PrinterStatus({ state, job }: PrinterStatusProps) {
  const { t } = useTranslation();

  return (
    <section>
      <h2>{t("printer.title")}</h2>

      <p>{t("printer.state", { state })}</p>

      {job && <p>{t("printer.job", { job: job.name })}</p>}
    </section>
  );
}
