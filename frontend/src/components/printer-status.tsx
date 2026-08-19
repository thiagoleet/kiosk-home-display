import type { PrintJob } from "../types/printer";
import type { PrinterState } from "../hooks/use-printer";

type PrinterStatusProps = {
  state: PrinterState;
  job: PrintJob | null;
};

export function PrinterStatus({ state, job }: PrinterStatusProps) {
  return (
    <section>
      <h2>Printer</h2>

      <p>State: {state}</p>

      {job && <p>Job: {job.name}</p>}
    </section>
  );
}
