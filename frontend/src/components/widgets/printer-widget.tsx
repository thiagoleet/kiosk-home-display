import { CheckCircle, LoaderCircle, Printer } from "lucide-react";

import type { PrintJob } from "../../types/printer";
import type { PrinterState } from "../../hooks/use-printer";

type PrinterWidgetProps = {
  state: PrinterState;
  job: PrintJob | null;
};

export function PrinterWidget({ state, job }: PrinterWidgetProps) {
  const isPrinting = state === "printing";

  return (
    <section className="widget printer-widget">
      <div className="widget-title">
        <Printer size={18} />
        <span>Printer</span>
      </div>

      <div className="printer-widget__state">
        {isPrinting ? (
          <>
            <LoaderCircle
              size={18}
              className="spin"
            />

            <span>Printing</span>
          </>
        ) : (
          <>
            <CheckCircle size={18} />

            <span>Ready</span>
          </>
        )}
      </div>

      <p className="printer-widget__job">{job?.name ?? "No active jobs"}</p>
    </section>
  );
}
