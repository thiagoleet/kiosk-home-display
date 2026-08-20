import { CheckCircle, LoaderCircle, Printer } from "lucide-react";

import { usePrinter } from "../../hooks/use-printer";
import { useTranslation } from "../../hooks/use-translation";

export function PrinterWidget() {
  const { state, currentJob } = usePrinter();
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
        {currentJob?.name ?? t("printer.noActiveJobs")}
      </p>
    </section>
  );
}
