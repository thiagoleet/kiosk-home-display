/**
 * The payload of the printer.started and printer.completed events.
 *
 * The daemon names the identifier `jobId`, not `id`: it is a CUPS job
 * identifier such as "Brother_DCP_1600_series-31", not a key of anything the
 * frontend holds.
 */
export type PrintJob = {
  jobId: string;
  name: string;
};
