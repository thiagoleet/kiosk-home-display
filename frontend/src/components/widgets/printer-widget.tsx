import { CheckCircle, LoaderCircle, Printer } from "lucide-react";

import type { PrintJob } from "../../types/printer";
import type { PrinterState } from "../../hooks/use-printer";
import { useTranslation } from "../../hooks/use-translation";

type PrinterWidgetProps = {
  state: PrinterState;
  job: PrintJob | null;
};

export function PrinterWidget({ state, job }: PrinterWidgetProps) {
  const isPrinting = state === "printing";
  const { t } = useTranslation();

  return (
    <section
      className="widget printer-widget"
      data-state={state}
    >
      <div className="widget-title">
        <Printer size={18} />
        <span>{t("printer.title")}</span>
      </div>

      <div className="printer-widget__state">
        {isPrinting ? (
          <>
            <LoaderCircle
              size={18}
              className="spin"
            />

            <span>{t("printer.printing")}</span>
          </>
        ) : (
          <>
            <CheckCircle size={18} />

            <span>{t("printer.ready")}</span>
          </>
        )}
      </div>

      <p className="printer-widget__job">
        {job?.name ?? t("printer.noActiveJobs")}
      </p>
    </section>
  );
}
